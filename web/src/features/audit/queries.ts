import { useQuery } from "@tanstack/react-query";
import { api, organizationHeaders } from "@/api/client";
import { scopeKey } from "@/features/workspace/public";
import type { Scope } from "@/features/workspace/public";

export function useAudit(scope: Scope, cursor?: string) {
  return useQuery({
    queryKey: [...scopeKey(scope), "audit", cursor ?? ""],
    queryFn: async ({ signal }) => {
      const { data } = await api.GET("/api/audit", {
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
        throw new Error("Audit response missing");
      }

      return data;
    },
  });
}
