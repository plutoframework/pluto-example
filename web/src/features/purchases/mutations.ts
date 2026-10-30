import { useCallback, useEffect, useRef, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { useActionMutation } from "@plutoframework/frontend/query";
import { ApiError } from "@/api/errors";
import { confirmPurchase } from "./cache";
import { api, organizationHeaders, writeHeaders } from "@/api/client";
import { scopeKey } from "@/features/workspace/public";
import type { Scope } from "@/features/workspace/public";
import type { CreatePurchase, Purchase } from "./types";
import { createDecisionFlow } from "./flows/decision";

type Command = {
  key: string;
  signal?: AbortSignal;
} & (
  | {
      action: "create";
      input: CreatePurchase;
    }
  | {
      action: "submit";
      id: string;
    }
  | {
      action: "decide";
      id: string;
      decision: "approved" | "rejected";
    }
);

export function usePurchaseCommand(scope: Scope) {
  const action = useActionMutation<Command, Purchase>({
    scope: JSON.stringify(scopeKey(scope)),
    classifyError: (error) =>
      error instanceof ApiError &&
      [400, 401, 403, 404, 422, 429].includes(error.status)
        ? "rejected"
        : "uncertain",
    execute: async (command, context) => {
      const signal = command.signal
        ? AbortSignal.any([command.signal, context.signal])
        : context.signal;

      const header = writeHeaders(scope.organization, command.key);

      if (command.action === "create") {
        const { data } = await api.POST("/api/purchases", {
          params: { header },
          signal,
          body: command.input,
        });

        if (!data) {
          throw new Error("Purchase response missing");
        }

        return data.data;
      }

      const params = {
        header,
        path: { purchaseID: command.id },
      };

      const { data } =
        command.action === "submit"
          ? await api.POST("/api/purchases/{purchaseID}/submit", {
            params,
            signal,
          })
          : await api.POST("/api/purchases/{purchaseID}/decision", {
            params,
            signal,
            body: { decision: command.decision },
          });

      if (!data) {
        throw new Error("Purchase response missing");
      }

      return data.data;
    },
    confirm: (purchase, _command, client) => {
      confirmPurchase(client, scope, purchase);
    },
    reconcile: (_command, client) =>
      client.invalidateQueries({ queryKey: scopeKey(scope) }),
  });

  const { run, get } = action;
  const mutateAsync = useCallback(
    async (command: Command) => {
      const resource = command.action === "create" ? "create" : command.id;
      const accepted = await run(resource, command);
      const result = get(resource);

      if (!accepted) {
        throw new Error("This action is pending or needs reconciliation.");
      }

      if (result?.status !== "succeeded" || !result.output) {
        throw result?.error ?? new Error("Action cancelled");
      }

      return result.output;
    },
    [run, get],
  );

  const latest = action.entries.at(-1);

  return {
    isPending: action.entries.some((entry) => entry.status === "pending"),
    isUncertain: action.entries.some((entry) => entry.status === "uncertain"),
    error: latest?.error instanceof Error ? latest.error : null,
    refreshError: latest?.refreshError !== undefined,
    reset() {
      for (const entry of action.entries) {
        void action.forget(entry.key);
      }
    },
    mutateAsync,
  };
}

export function usePurchaseExport(scope: Scope) {
  return useMutation({
    retry: false,
    mutationFn: async () => {
      const { data } = await api.GET("/api/purchases/export", {
        params: { header: organizationHeaders(scope.organization) },
      });

      if (!data) {
        throw new Error("Export response missing");
      }

      return data.data;
    },
  });
}

export function usePurchaseReview(scope: Scope) {
  return useMutation({
    retry: false,
    mutationFn: async ({ id, signal }: {
      id: string;
      signal: AbortSignal;
    }) => {
      const { data } = await api.POST("/api/purchases/{purchaseID}/review", {
        signal,
        params: {
          header: organizationHeaders(scope.organization),
          path: { purchaseID: id },
        },
      });

      if (!data) {
        throw new Error("Review response missing");
      }

      return data.data;
    },
  });
}

export function useReviewController(scope: Scope) {
  const review = usePurchaseReview(scope);
  const controller = useRef<AbortController | null>(null);
  useEffect(() => () => controller.current?.abort(), []);

  return {
    review,
    controller,
  };
}

// Constructors are pure, but controller ownership follows committed lifecycles.
// Each StrictMode setup creates a fresh flow; cleanup permanently fences its work.
export function usePurchaseDecision(scope: Scope, id: string) {
  const { mutateAsync } = usePurchaseCommand(scope);
  const identity = JSON.stringify([...scopeKey(scope), id]);
  const [owned, setOwned] = useState<{
    identity: string;
    flow: ReturnType<typeof createDecisionFlow>;
  } | null>(null);

  useEffect(() => {
    const flow = createDecisionFlow((verdict, key, signal) =>
      mutateAsync({
        action: "decide",
        id,
        decision: verdict,
        key,
        signal,
      }),
    );

    setOwned({
      identity,
      flow,
    });

    return () => flow.dispose();
  }, [identity, id, mutateAsync]);

  return owned?.identity === identity ? owned.flow : null;
}
