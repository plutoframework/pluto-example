import { useSearch, useNavigate } from "@tanstack/react-router";
import { useWorkspace } from "@/features/workspace/public";
import { errorMessage } from "@/api/errors";
import { PageHeader, Loading, Notice, EmptyState } from "@/shared/ui/feedback";
import { Pagination } from "@/shared/ui/Pagination";
import { date, initials } from "@/shared/lib/format";
import { nextPage, previousPage } from "@/shared/lib/search";
import { useAudit } from "../queries";
export function ActivityPage() {
  const { scope } = useWorkspace();
  const search = useSearch({ from: "/activity" });
  const navigate = useNavigate({ from: "/activity" });
  const query = useAudit(scope, search.cursor);
  return (
    <>
      <PageHeader
        eyebrow="WORKSPACE / ACTIVITY"
        title="Every decision has a story."
        description="A server-recorded history of successful actions within your organization access."
      />
      <section className="panel">
        <div className="panel-heading">
          <div>
            <h2>Activity log</h2>
            <p>
              Members see their own actions. Reviewers see organization history.
            </p>
          </div>
          <span className="tag">AUDITED BY PLUTO</span>
        </div>
        {query.isPending ? (
          <Loading label="Loading activity history…" />
        ) : query.error ? (
          <Notice tone="error">{errorMessage(query.error)}</Notice>
        ) : !query.data?.data.length ? (
          <EmptyState
            title="A clean slate"
            description="Successful requests and role changes will appear here."
          />
        ) : (
          <ol className="activity-list">
            {query.data.data.map((event) => (
              <li key={event.id}>
                <span className="activity-marker" aria-hidden="true">
                  ↗
                </span>
                <div className="activity-copy">
                  <h3>{event.action.replaceAll(/[._]/g, " ")}</h3>
                  <p>
                    <span className="avatar tiny">
                      {initials(event.actor_id)}
                    </span>
                    {event.actor_id === scope.actor
                      ? "You"
                      : "Organization member"}
                    <span className="separator">/</span>
                    <code>{event.resource_id}</code>
                  </p>
                </div>
                <div className="activity-time">
                  <span className="status status-approved">
                    {event.outcome}
                  </span>
                  <time dateTime={event.occurred_at}>
                    {date(event.occurred_at)}
                  </time>
                </div>
              </li>
            ))}
          </ol>
        )}
        <Pagination
          count={query.data?.data.length ?? 0}
          previous={!!search.trail?.length}
          more={query.data?.page.has_more ?? false}
          onPrevious={() => {
            void navigate({ search: (old) => previousPage(old) });
          }}
          onNext={() => {
            if (query.data)
              void navigate({
                search: (old) => nextPage(old, query.data!.page.next_cursor),
              });
          }}
        />
      </section>
    </>
  );
}
