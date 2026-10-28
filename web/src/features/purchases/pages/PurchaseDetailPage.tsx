import { useRef } from "react";
import { Link, useParams } from "@tanstack/react-router";
import { useWorkspace } from "@/features/workspace/public";
import { errorMessage } from "@/api/errors";
import { PageHeader, Loading, Notice } from "@/shared/ui/feedback";
import { date, money } from "@/shared/lib/format";
import { usePurchase } from "../queries";
import { usePurchaseCommand } from "../mutations";
import { PurchaseStatus } from "../components/PurchaseStatus";

export function PurchaseDetailPage() {
  const { purchaseId } = useParams({ from: "/purchases/$purchaseId" });
  const { scope, can } = useWorkspace();
  const query = usePurchase(scope, purchaseId);
  const command = usePurchaseCommand(scope);
  const keys = useRef(new Map<string, string>());
  if (query.isPending) return <Loading label="Opening purchase request…" />;
  if (query.error || !query.data)
    return <Notice tone="error">{errorMessage(query.error)}</Notice>;
  const record = query.data;
  const own = record.creator_id === scope.actor;
  const act = async (action: "submit" | "approved" | "rejected") => {
    if (!keys.current.has(action))
      keys.current.set(action, crypto.randomUUID());
    const key = keys.current.get(action)!;
    try {
      if (action === "submit")
        await command.mutateAsync({ action: "submit", id: record.id, key });
      else
        await command.mutateAsync({
          action: "decide",
          id: record.id,
          decision: action,
          key,
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
        <div className="panel detail-card">
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
        </div>
        <div className="panel detail-card">
          <span className="eyebrow">NEXT STEP</span>
          <h2>
            {record.status === "draft"
              ? "Ready when you are."
              : record.status === "submitted"
                ? "A considered decision."
                : "Decision recorded."}
          </h2>
          <p className="muted">
            {record.status === "draft"
              ? "Submit your draft for another team member to review."
              : record.status === "submitted"
                ? "An authorized reviewer can approve or decline this request. Creators cannot review their own requests."
                : "This request has reached its final state. Its history remains in the activity log."}
          </p>
          {command.error && (
            <Notice tone="error">{errorMessage(command.error)}</Notice>
          )}
          <div className="button-row">
            {record.status === "draft" && own && can("purchases:create") && (
              <button
                className="button primary"
                disabled={command.isPending}
                onClick={() => {
                  void act("submit");
                }}
              >
                Submit for review ↗
              </button>
            )}
            {record.status === "submitted" &&
              !own &&
              can("purchases:decide") && (
                <>
                  <button
                    className="button primary"
                    disabled={command.isPending}
                    onClick={() => {
                      void act("approved");
                    }}
                  >
                    Approve request
                  </button>
                  <button
                    className="button danger"
                    disabled={command.isPending}
                    onClick={() => {
                      void act("rejected");
                    }}
                  >
                    Decline request
                  </button>
                </>
              )}
          </div>
        </div>
      </section>
    </>
  );
}
