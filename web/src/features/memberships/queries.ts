import { useQuery } from "@tanstack/react-query";
import { api, organizationHeaders } from "@/api/client";
import { scopeKey } from "@/features/workspace/public";
import type { Scope } from "@/features/workspace/public";

export function useMemberships(
  scope: Scope,
  cursor: string | undefined,
  enabled: boolean,
) {
  return useQuery({
    queryKey: [...scopeKey(scope), "memberships", cursor ?? ""],
    enabled,
    queryFn: async ({ signal }) => {
      const { data } = await api.GET("/api/memberships", {
        signal,
        params: {
          header: organizationHeaders(scope.organization),
          query: {
            limit: 20,
            cursor,
          },
        },
      });

      if (!data) {
        throw new Error("Membership response missing");
      }

      return data;
    },
  });
}
