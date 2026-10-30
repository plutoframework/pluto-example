import { Button } from "@heroui/react";
import { useMachine } from "@plutoframework/frontend/react";
import { Modal } from "@/shared/ui/Modal";
import { Notice } from "@/shared/ui/feedback";
import type { createDecisionFlow, DecisionEvent } from "../flows/decision";

export function PurchaseDecisionControls({
  flow,
}: {
  flow: ReturnType<typeof createDecisionFlow>;
}) {
  const { state } = useMachine(flow);
  const send = (event: DecisionEvent) => {

    // Expected API errors are flow events; reset if an unexpected contract fails.
    void flow.send(event).catch(() => flow.reset());
  };

  if (state.status === "submitting") {
    return <Button isDisabled>Recording decision…</Button>;
  }

  if (state.status === "succeeded") {
    return (
      <Notice tone="success">
        Decision recorded. Purchase details are refreshing.
      </Notice>
    );
  }

  if (state.status === "failed") {
    return (
      <div>
        <Notice tone="error">
          {state.message} Refresh the purchase before retrying an uncertain
          request.
        </Notice>
        <Button onClick={() => send({ type: "dismiss" })}>Dismiss</Button>
      </div>
    );
  }

  return (
    <>
      <Button
        variant="primary"
        onClick={() =>
          send({
            type: "choose",
            verdict: "approved",
          })
        }
      >
        Approve request
      </Button>
      <Button
        variant="danger-soft"
        onClick={() =>
          send({
            type: "choose",
            verdict: "rejected",
          })
        }
      >
        Decline request
      </Button>
      {state.status === "confirming" && (
        <Modal
          title={
            state.verdict === "approved"
              ? "Approve this request?"
              : "Decline this request?"
          }
          onClose={() => send({ type: "cancel" })}
        >
          <p>
            This decision will be recorded in the organization’s activity log.
          </p>
          <div className="button-row">
            <Button
              variant="primary"
              onClick={() =>
                send({
                  type: "confirm",
                  requestId: crypto.randomUUID(),
                })
              }
            >
              {state.verdict === "approved"
                ? "Confirm approval"
                : "Confirm decline"}
            </Button>
            <Button
              variant="secondary"
              onClick={() => send({ type: "cancel" })}
            >
              Cancel
            </Button>
          </div>
        </Modal>
      )}
    </>
  );
}
