import { useQuery } from "@tanstack/react-query";
import { api, organizationHeaders } from "@/api/client";
import { scopeKey } from "./types";
import type { Scope } from "./types";

export function useWorkspaceQueries(scope: Scope) {
  const params = { header: organizationHeaders(scope.organization) };
  const access = useQuery({
    queryKey: [...scopeKey(scope), "access"],
    staleTime: 0,
    retry: false,
    queryFn: async ({ signal }) => {
      const { data } = await api.GET("/api/access", { params, signal });
      if (
        !data ||
        data.data.actor_id !== scope.actor ||
        data.data.organization_id !== scope.organization
      )
        throw new Error("Workspace identity changed; refresh your session");
      return data.data;
    },
  });
  const organization = useQuery({
    queryKey: [...scopeKey(scope), "organization"],
    enabled: !!access.data,
    queryFn: async ({ signal }) => {
      const { data } = await api.GET("/api/organization", { params, signal });
      if (!data || data.data.id !== scope.organization)
        throw new Error("Workspace scope changed");
      return data.data;
    },
  });
  const features = useQuery({
    queryKey: [...scopeKey(scope), "features"],
    enabled: !!access.data,
    queryFn: async ({ signal }) => {
      const { data } = await api.GET("/api/features", { params, signal });
      if (!data) throw new Error("Feature response missing");
      return data.data;
    },
  });
  return { access, organization, features };
}
