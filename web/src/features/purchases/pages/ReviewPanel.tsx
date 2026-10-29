import { Button, Card } from "@heroui/react";

import type { Scope } from "@/features/workspace/public";
import { Notice } from "@/shared/ui/feedback";
import { errorMessage } from "@/api/errors";
import { useReviewController } from "../mutations";

export function ReviewPanel({ scope, id }: { scope: Scope; id: string }) {
  const { review, controller } = useReviewController(scope);
  return (
    <Card className="panel detail-card">
      <h2>Read-only purchase review</h2>
      <p className="muted">
        Suggestions for a human decision. This review cannot approve, pay, or
        change your request.
      </p>
      <div className="button-row">
        <Button
          isDisabled={review.isPending}
          onClick={() => {
            controller.current = new AbortController();
            review.mutate({ id, signal: controller.current.signal });
          }}
        >
          {review.isPending ? "Reviewing…" : "Review purchase"}
        </Button>
        {review.isPending && (
          <Button
            variant="secondary"
            onClick={() => controller.current?.abort()}
          >
            Cancel review
          </Button>
        )}
      </div>
      {review.error && (
        <Notice tone="error">{errorMessage(review.error)}</Notice>
      )}
      {review.data && (
        <div aria-live="polite">
          <strong>
            {review.data.demo
              ? "Demo · deterministic suggestions"
              : "AI-generated suggestions · verify before acting"}
          </strong>
          <p>{review.data.summary}</p>
          <ul>
            {review.data.next_steps.map((step, index) => (
              <li key={index}>{step}</li>
            ))}
          </ul>
        </div>
      )}
    </Card>
  );
}
