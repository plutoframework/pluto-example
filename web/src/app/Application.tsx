import { Outlet } from "@tanstack/react-router";
import { useSession, AuthPage } from "@/features/session/public";
import { WorkspaceProvider, useWorkspace } from "@/features/workspace/public";
import { Loading, Notice, EmptyState } from "@/shared/ui/feedback";
import { errorMessage } from "@/api/errors";
import { AppShell } from "./AppShell";

function WorkspaceContent() {
  const { pending, error } = useWorkspace();
  const { session, refresh } = useSession();
  if (pending) return <Loading />;
  if (error)
    return (
      <EmptyState
        title="Let's find your workspace."
        description={errorMessage(error)}
      >
        <p>
          A verified account does not automatically receive organization
          membership. A trusted developer must explicitly assign access for this
          example.
        </p>
        <label className="account-reference">
          Your account ID
          <input
            readOnly
            value={session?.actor_id ?? ""}
            onFocus={(event) => event.currentTarget.select()}
          />
        </label>
        <button className="button ghost" onClick={refresh}>
          Refresh session
        </button>
      </EmptyState>
    );
  return <Outlet />;
}
export function Application() {
  const { session, pending, error, refresh } = useSession();
  if (pending)
    return (
      <main className="full-loading">
        <Loading label="Finding your session…" />
      </main>
    );
  if (error)
    return (
      <main className="full-loading">
        <Notice tone="error">
          {errorMessage(error)}{" "}
          <button className="text-button" onClick={refresh}>
            Try again
          </button>
        </Notice>
      </main>
    );
  if (!session) return <AuthPage />;
  return (
    <WorkspaceProvider
      key={`${session.actor_id}:${session.id}`}
      session={session}
    >
      <AppShell>
        <WorkspaceContent />
      </AppShell>
    </WorkspaceProvider>
  );
}
