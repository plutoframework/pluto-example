import { createContext, useContext, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { Session } from "@/features/session/public";
import { useWorkspaceQueries } from "./queries";
import type { Access, Organization, Scope } from "./types";

type WorkspaceState = {
  scope: Scope;
  access?: Access;
  organization?: Organization;
  features: Record<string, boolean>;
  pending: boolean;
  error: unknown;
  selectOrganization: (id: string) => Promise<void>;
  can: (permission: string) => boolean;
};

const WorkspaceContext = createContext<WorkspaceState | null>(null);
export function WorkspaceProvider({
  session,
  children,
}: {
  session: Session;
  children: ReactNode;
}) {
  const [organization, setOrganization] = useState("northstar");
  const client = useQueryClient();
  const scope = useMemo(
    () => ({
      actor: session.actor_id,
      session: session.id,
      organization,
    }),
    [session.actor_id, session.id, organization],
  );

  const queries = useWorkspaceQueries(scope);
  const selectOrganization = async (id: string) => {
    if (id === organization) {
      return;
    }

    await client.cancelQueries({ queryKey: ["scope"] });
    client.removeQueries({ queryKey: ["scope"] });
    client.getMutationCache().clear();
    setOrganization(id);
  };

  return (
    <WorkspaceContext.Provider
      value={{
        scope,
        access: queries.access.data,
        organization: queries.organization.data,
        features: queries.features.data ?? {},
        pending:
          queries.access.isPending ||
          (!!queries.access.data && queries.organization.isPending),
        error:
          queries.access.error ??
          queries.organization.error ??
          queries.features.error,
        selectOrganization,
        can: (permission) =>
          queries.access.data?.permissions.includes(permission) ?? false,
      }}
    >
      <div key={organization} className="workspace-root">
        {children}
      </div>
    </WorkspaceContext.Provider>
  );
}

export function useWorkspace(): WorkspaceState {
  const value = useContext(WorkspaceContext);

  if (!value) {
    throw new Error("WorkspaceProvider is required");
  }

  return value;
}
