import { useRef, useState } from "react";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useWorkspace, useScopeGuard } from "@/features/workspace/public";
import { errorMessage } from "@/api/errors";
import { PageHeader, Loading, Notice, EmptyState } from "@/shared/ui/feedback";
import { Pagination } from "@/shared/ui/Pagination";
import { Modal } from "@/shared/ui/Modal";
import { money } from "@/shared/lib/format";
import { nextPage, previousPage } from "@/shared/lib/search";
import { downloadJson } from "@/shared/lib/download";
import { usePurchases } from "../queries";
import { usePurchaseCommand, usePurchaseExport } from "../mutations";
import { PurchaseTable } from "../components/PurchaseTable";
import { CreatePurchaseForm } from "../forms/CreatePurchaseForm";
import { statusLabels } from "../types";
import type { PurchaseStatus } from "../types";

export function PurchaseListPage() {
  const { scope, can, features } = useWorkspace();
  const isCurrentScope = useScopeGuard(scope);
  const search = useSearch({ from: "/" });
  const navigate = useNavigate({ from: "/" });
  const query = usePurchases(scope, search.cursor);
  const command = usePurchaseCommand(scope);
  const exporter = usePurchaseExport(scope);
  const [creating, setCreating] = useState(false);
  const intent = useRef<{ signature: string; key: string } | null>(null);
  const records = query.data?.data ?? [];
  const visible = records.filter(
    (record) =>
      (!search.status || record.status === search.status) &&
      (!search.q ||
        record.title.toLowerCase().includes(search.q.toLowerCase())),
  );
  const submitted = records.filter(
    (record) => record.status === "submitted",
  ).length;
  const drafts = records.filter((record) => record.status === "draft").length;
  return (
    <>
      <PageHeader
        eyebrow="WORKSPACE / PURCHASING"
        title="Make good things happen."
        description="A clear path from your team's next idea to its next delivery."
        action={
          can("purchases:create") && (
            <button
              className="button primary"
              onClick={() => {
                intent.current = null;
                command.reset();
                setCreating(true);
              }}
            >
              <span aria-hidden="true">＋</span>New request
            </button>
          )
        }
      />
      <div className="overview-strip">
        <div>
          <span className="eyebrow">AWAITING REVIEW</span>
          <strong>{submitted.toString().padStart(2, "0")}</strong>
          <span>Ready for a considered decision</span>
        </div>
        <div>
          <span className="eyebrow">IN DRAFT</span>
          <strong>{drafts.toString().padStart(2, "0")}</strong>
          <span>Good ideas taking shape</span>
        </div>
        <div className="overview-value">
          <span className="eyebrow">VALUE ON THIS PAGE</span>
          <strong>
            {money(
              records.reduce((sum, record) => sum + record.amount_cents, 0),
            )}
          </strong>
          <span>USD · {records.length} visible records loaded</span>
        </div>
      </div>
      <section className="panel">
        <div className="panel-heading">
          <div>
            <h2>
              Purchase requests{" "}
              <span className="count-tag">{records.length}</span>
            </h2>
            <p>Only requests your current membership allows you to see.</p>
          </div>
          {can("purchases:export") && features.purchase_export && (
            <button
              className="button ghost small"
              disabled={exporter.isPending}
              onClick={async () => {
                try {
                  const exported = await exporter.mutateAsync();
                  if (!isCurrentScope()) return;
                  downloadJson(
                    exported,
                    `${scope.organization}-purchases.json`,
                  );
                } catch {
                  /* Error shown below. */
                }
              }}
            >
              Export JSON ↗
            </button>
          )}
        </div>
        {exporter.error && (
          <Notice tone="error">{errorMessage(exporter.error)}</Notice>
        )}
        <div className="filters">
          <label className="search-field">
            <span className="sr-only">Search this page</span>
            <span aria-hidden="true">⌕</span>
            <input
              type="search"
              placeholder="Search this page…"
              value={search.q ?? ""}
              onChange={(event) => {
                void navigate({
                  search: (old) => ({
                    ...old,
                    q: event.target.value || undefined,
                  }),
                  replace: true,
                });
              }}
            />
          </label>
          <label className="filter-select">
            <span>Status</span>
            <select
              aria-label="Filter status on this page"
              value={search.status ?? ""}
              onChange={(event) => {
                void navigate({
                  search: (old) => ({
                    ...old,
                    status: (event.target.value as PurchaseStatus) || undefined,
                  }),
                  replace: true,
                });
              }}
            >
              <option value="">All statuses</option>
              {Object.entries(statusLabels).map(([value, label]) => (
                <option key={value} value={value}>
                  {label}
                </option>
              ))}
            </select>
          </label>
        </div>
        {query.isPending ? (
          <Loading label="Loading purchase requests…" />
        ) : query.error ? (
          <Notice tone="error">
            {errorMessage(query.error)}{" "}
            <button
              className="text-button"
              onClick={() => {
                void query.refetch();
              }}
            >
              Try again
            </button>
          </Notice>
        ) : visible.length ? (
          <PurchaseTable purchases={visible} actor={scope.actor} />
        ) : (
          <EmptyState
            title={
              records.length
                ? "No matches on this page"
                : "Your next idea starts here"
            }
            description={
              records.length
                ? "Try another search or status, or move to another page."
                : "Create a draft request and give your team something to move forward."
            }
          />
        )}
        <Pagination
          count={visible.length}
          previous={!!search.trail?.length}
          more={query.data?.page.has_more ?? false}
          onPrevious={() => {
            void navigate({
              search: (old) => ({ ...old, ...previousPage(old) }),
            });
          }}
          onNext={() => {
            if (query.data)
              void navigate({
                search: (old) => ({
                  ...old,
                  ...nextPage(old, query.data!.page.next_cursor),
                }),
              });
          }}
        />
      </section>
      <p className="page-footnote">
        <span className="signal-dot" />
        Permissions are checked by Pluto on every request. Search and status
        filters apply to the loaded page.
      </p>
      {creating && (
        <Modal
          title="A new purchase request"
          onClose={() => setCreating(false)}
        >
          {command.error && (
            <Notice tone="error">{errorMessage(command.error)}</Notice>
          )}
          <CreatePurchaseForm
            pending={command.isPending}
            onCancel={() => setCreating(false)}
            onSubmit={async (input) => {
              const signature = JSON.stringify(input);
              if (intent.current?.signature !== signature)
                intent.current = { signature, key: crypto.randomUUID() };
              try {
                const record = await command.mutateAsync({
                  action: "create",
                  input,
                  key: intent.current.key,
                });
                if (!isCurrentScope()) return;
                setCreating(false);
                void navigate({
                  to: "/purchases/$purchaseId",
                  params: { purchaseId: record.id },
                });
              } catch {
                /* Preserve this logical action's key for a retry of the same input. */
              }
            }}
          />
        </Modal>
      )}
    </>
  );
}
