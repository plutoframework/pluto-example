import { Button, Input, Label, TextField } from "@heroui/react";
import { Outlet } from "@tanstack/react-router";
import { useSession, AuthPage } from "@/features/session/public";
import { WorkspaceProvider, useWorkspace } from "@/features/workspace/public";
import { Loading, Notice, EmptyState } from "@/shared/ui/feedback";
import { errorMessage } from "@/api/errors";
import { AppShell } from "./AppShell";

function WorkspaceContent() {
  const { pending, error } = useWorkspace();
  const { session, refresh } = useSession();

  if (pending) {
    return <Loading />;
  }

  if (error) {
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
        <TextField
          className="account-reference"
          isReadOnly
          value={session?.actor_id ?? ""}
        >
          <Label>Your account ID</Label>
          <Input
            variant="secondary"
            onFocus={(event) => event.currentTarget.select()}
          />
        </TextField>
        <Button variant="outline" onClick={refresh}>
          Refresh session
        </Button>
      </EmptyState>
    );
  }

  return <Outlet />;
}

export function Application() {
  const { session, pending, error, refresh } = useSession();

  if (pending) {
    return (
      <main className="full-loading">
        <Loading label="Finding your session…" />
      </main>
    );
  }

  if (error) {
    return (
      <main className="full-loading">
        <Notice tone="error">
          {errorMessage(error)}{" "}
          <Button
            className="text-button"
            variant="ghost"
            size="sm"
            onClick={refresh}
          >
            Try again
          </Button>
        </Notice>
      </main>
    );
  }

  if (!session) {
    return <AuthPage />;
  }

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
