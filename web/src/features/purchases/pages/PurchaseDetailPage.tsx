import { Button, Card } from "@heroui/react";
import { useRef } from "react";
import { Link, useParams } from "@tanstack/react-router";
import { useWorkspace } from "@/features/workspace/public";
import { errorMessage } from "@/api/errors";
import { PageHeader, Loading, Notice } from "@/shared/ui/feedback";
import { date, money } from "@/shared/lib/format";
import { usePurchase } from "../queries";
import { usePurchaseCommand } from "../mutations";
import { ReviewPanel } from "./ReviewPanel";
import { PurchaseDecision } from "./PurchaseDecision";
import { PurchaseStatus } from "../components/PurchaseStatus";

export function PurchaseDetailPage() {
  const { purchaseId } = useParams({ from: "/purchases/$purchaseId" });
  const { scope, can, features } = useWorkspace();
  const query = usePurchase(scope, purchaseId);
  const command = usePurchaseCommand(scope);
  const keys = useRef(new Map<string, string>());

  if (query.isPending) {
    return <Loading label="Opening purchase request…" />;
  }

  if (query.error || !query.data) {
    return <Notice tone="error">{errorMessage(query.error)}</Notice>;
  }

  const record = query.data;
  const nextSteps: Record<
    string,
    {
      title: string;
      description: string;
    }
  > = {
    draft: {
      title: "Ready when you are.",
      description: "Submit your draft for another team member to review.",
    },
    submitted: {
      title: "A considered decision.",
      description:
        "An authorized reviewer can approve or decline this request. Creators cannot review their own requests.",
    },
  };

  const nextStep = nextSteps[record.status] ?? {
    title: "Decision recorded.",
    description:
      "This request has reached its final state. Its history remains in the activity log.",
  };

  const own = record.creator_id === scope.actor;
  const submit = async () => {
    if (!keys.current.has("submit")) {
      keys.current.set("submit", crypto.randomUUID());
    }

    try {
      await command.mutateAsync({
        action: "submit",
        id: record.id,
        key: keys.current.get("submit")!,
      });
    } catch {

      /* Display sanitized error and refresh canonical server state. */
    }
  };

  return (
    <>
      <Link className="back-link" to="/">
        ← All purchase requests
      </Link>
      <PageHeader
        eyebrow="PURCHASING / REQUEST DETAILS"
        title={record.title}
        description={`Created ${date(record.created_at)} · ${record.id}`}
        action={<PurchaseStatus status={record.status} />}
      />
      <section className="detail-grid">
        <Card className="panel detail-card">
          <span className="eyebrow">ESTIMATED VALUE</span>
          <h2 className="detail-amount">{money(record.amount_cents)}</h2>
          <dl>
            <div>
              <dt>Requested by</dt>
              <dd>
                {own ? "You" : "Organization member"}
                <small>{record.creator_id}</small>
              </dd>
            </div>
            <div>
              <dt>Organization</dt>
              <dd>{scope.organization}</dd>
            </div>
            <div>
              <dt>Last updated</dt>
              <dd>{date(record.updated_at)}</dd>
            </div>
            <div>
              <dt>Record version</dt>
              <dd>{record.version}</dd>
            </div>
          </dl>
        </Card>
        <Card className="panel detail-card">
          <span className="eyebrow">NEXT STEP</span>
          <h2>{nextStep.title}</h2>
          <p className="muted">{nextStep.description}</p>
          {command.error && (
            <Notice tone="error">{errorMessage(command.error)}</Notice>
          )}
          {command.isPending && <p role="status">Submitting your request…</p>}
          {command.isUncertain && (
            <Notice tone="error">
              The request outcome is unknown. Refresh to verify its status
              before retrying.
            </Notice>
          )}
          {command.refreshError && (
            <Notice tone="error">
              The write completed, but current details could not refresh.
              Refresh this page.
            </Notice>
          )}
          <div className="button-row">
            {record.status === "draft" && own && can("purchases:create") && (
              <Button
                variant="primary"
                isDisabled={command.isPending || command.isUncertain}
                onClick={() => {
                  void submit();
                }}
              >
                {command.isPending ? "Submitting…" : "Submit for review ↗"}
              </Button>
            )}
            {record.status === "submitted" &&
              !own &&
              can("purchases:decide") && (
                <PurchaseDecision
                  key={`${scope.actor}:${scope.session}:${scope.organization}:${record.id}`}
                  scope={scope}
                  id={record.id}
                />
              )}
          </div>
        </Card>
      </section>
      {features.purchase_review && can("purchases:read") && (
        <ReviewPanel
          key={`${scope.organization}:${scope.actor}:${scope.session}:${purchaseId}`}
          scope={scope}
          id={purchaseId}
        />
      )}
    </>
  );
}
