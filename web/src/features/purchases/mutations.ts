import { useEffect, useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, organizationHeaders, writeHeaders } from "@/api/client";
import { scopeKey } from "@/features/workspace/public";
import type { Scope } from "@/features/workspace/public";
import type { CreatePurchase } from "./types";
import { createDecisionFlow } from "./flows/decision";

type Command = { key: string; signal?: AbortSignal } & (
  | { action: "create"; input: CreatePurchase }
  | { action: "submit"; id: string }
  | { action: "decide"; id: string; decision: "approved" | "rejected" }
);
export function usePurchaseCommand(scope: Scope) {
  const client = useQueryClient();
  return useMutation({
    retry: false,
    mutationFn: async (command: Command) => {
      const header = writeHeaders(scope.organization, command.key);
      if (command.action === "create") {
        const { data } = await api.POST("/api/purchases", {
          params: { header },
          signal: command.signal,
          body: command.input,
        });
        if (!data) throw new Error("Purchase response missing");
        return data.data;
      }
      const params = { header, path: { purchaseID: command.id } };
      const { data } =
        command.action === "submit"
          ? await api.POST("/api/purchases/{purchaseID}/submit", {
              params,
              signal: command.signal,
            })
          : await api.POST("/api/purchases/{purchaseID}/decision", {
              params,
              signal: command.signal,
              body: { decision: command.decision },
            });
      if (!data) throw new Error("Purchase response missing");
      return data.data;
    },
    onSettled: async () => {
      await client.invalidateQueries({ queryKey: scopeKey(scope) });
    },
  });
}
export function usePurchaseExport(scope: Scope) {
  return useMutation({
    retry: false,
    mutationFn: async () => {
      const { data } = await api.GET("/api/purchases/export", {
        params: { header: organizationHeaders(scope.organization) },
      });
      if (!data) throw new Error("Export response missing");
      return data.data;
    },
  });
}

export function usePurchaseReview(scope: Scope) {
  return useMutation({
    retry: false,
    mutationFn: async ({ id, signal }: { id: string; signal: AbortSignal }) => {
      const { data } = await api.POST("/api/purchases/{purchaseID}/review", {
        signal,
        params: {
          header: organizationHeaders(scope.organization),
          path: { purchaseID: id },
        },
      });
      if (!data) throw new Error("Review response missing");
      return data.data;
    },
  });
}

export function useReviewController(scope: Scope) {
  const review = usePurchaseReview(scope);
  const controller = useRef<AbortController | null>(null);
  useEffect(() => () => controller.current?.abort(), []);
  return { review, controller };
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
      mutateAsync({ action: "decide", id, decision: verdict, key, signal }),
    );
    setOwned({ identity, flow });
    return () => flow.dispose();
  }, [identity, id, mutateAsync]);
  return owned?.identity === identity ? owned.flow : null;
}
