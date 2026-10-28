import { useEffect, useMemo } from "react";
import type { Scope } from "./types";

// Async UI effects must not navigate, download or display results after their
// account/organization screen has been replaced. This never authorizes a write.
export function useScopeGuard(scope: Scope): () => boolean {
  const token = useMemo(
    () => ({
      active: true,
      actor: scope.actor,
      session: scope.session,
      organization: scope.organization,
    }),
    [scope.actor, scope.session, scope.organization],
  );
  useEffect(() => {
    token.active = true;
    return () => {
      token.active = false;
    };
  }, [token]);
  return () => token.active;
}
