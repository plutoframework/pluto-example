import { useQuery } from "@tanstack/react-query";
import { api, organizationHeaders } from "@/api/client";
import { scopeKey } from "@/features/workspace/public";
import type { Scope } from "@/features/workspace/public";

export function usePurchases(scope: Scope, cursor?: string) {
  return useQuery({
    queryKey: [...scopeKey(scope), "purchases", cursor ?? ""],
    queryFn: async ({ signal }) => {
      const { data } = await api.GET("/api/purchases", {
        signal,
        params: {
          header: organizationHeaders(scope.organization),
          query: { limit: 20, cursor },
        },
      });
      if (!data) throw new Error("Purchase page missing");
      return data;
    },
  });
}
export function usePurchase(scope: Scope, id: string) {
  return useQuery({
    queryKey: [...scopeKey(scope), "purchase", id],
    queryFn: async ({ signal }) => {
      const { data } = await api.GET("/api/purchases/{purchaseID}", {
        signal,
        params: {
          header: organizationHeaders(scope.organization),
          path: { purchaseID: id },
        },
      });
      if (!data || data.data.organization_id !== scope.organization)
        throw new Error("Purchase scope changed");
      return data.data;
    },
  });
}
