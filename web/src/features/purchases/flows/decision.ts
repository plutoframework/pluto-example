import { createMachine } from "@plutoframework/frontend/state";
import type { Decision } from "@plutoframework/frontend/state";
import { errorMessage } from "@/api/errors";

export type Verdict = "approved" | "rejected";
export type DecisionState =
  | { readonly status: "viewing" }
  | { readonly status: "confirming"; readonly verdict: Verdict }
  | {
      readonly status: "submitting";
      readonly verdict: Verdict;
      readonly requestId: string;
    }
  | { readonly status: "succeeded"; readonly verdict: Verdict }
  | { readonly status: "failed"; readonly message: string };
export type DecisionEvent =
  | { readonly type: "choose"; readonly verdict: Verdict }
  | { readonly type: "cancel" }
  | { readonly type: "confirm"; readonly requestId: string }
  | { readonly type: "resolved"; readonly requestId: string }
  | {
      readonly type: "failed";
      readonly requestId: string;
      readonly message: string;
    }
  | { readonly type: "dismiss" };
type Effect = { readonly verdict: Verdict; readonly requestId: string };

export function decisionTransition(
  current: DecisionState,
  event: DecisionEvent,
): Decision<DecisionState, Effect> {
  switch (current.status) {
    case "viewing":
      if (event.type === "choose") {
        return {
          accepted: true,
          state: { status: "confirming", verdict: event.verdict },
        };
      }
      break;
    case "confirming":
      if (event.type === "cancel")
        return { accepted: true, state: { status: "viewing" } };
      if (event.type === "confirm" && event.requestId !== "") {
        return {
          accepted: true,
          state: {
            status: "submitting",
            verdict: current.verdict,
            requestId: event.requestId,
          },
          effect: { verdict: current.verdict, requestId: event.requestId },
        };
      }
      break;
    case "submitting":
      if (event.type === "resolved" && event.requestId === current.requestId) {
        return {
          accepted: true,
          state: { status: "succeeded", verdict: current.verdict },
        };
      }
      if (event.type === "failed" && event.requestId === current.requestId) {
        return {
          accepted: true,
          state: { status: "failed", message: event.message },
        };
      }
      break;
    case "failed":
      if (event.type === "dismiss")
        return { accepted: true, state: { status: "viewing" } };
      break;
    case "succeeded":
      break;
  }
  return { accepted: false, state: current };
}

export function createDecisionFlow(
  decide: (
    verdict: Verdict,
    key: string,
    signal: AbortSignal,
  ) => Promise<unknown>,
  newKey: () => string = () => crypto.randomUUID(),
) {
  // Retain the same business key on retries; UI correlation IDs are separate.
  const keys = new Map<Verdict, string>();
  return createMachine<DecisionState, DecisionEvent, Effect>({
    initial: { status: "viewing" },
    transition: decisionTransition,
    async execute(effect, { signal, send }) {
      try {
        let key = keys.get(effect.verdict);
        if (!key) {
          key = newKey();
          keys.set(effect.verdict, key);
        }
        await decide(effect.verdict, key, signal);
      } catch (error) {
        await send({
          type: "failed",
          requestId: effect.requestId,
          message: errorMessage(error),
        });
        return;
      }
      await send({ type: "resolved", requestId: effect.requestId });
    },
  });
}
