import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, writeHeaders } from "@/api/client";
import { scopeKey } from "@/features/workspace/public";
import type { Scope } from "@/features/workspace/public";
export function useChangeRole(scope: Scope) {
  const client = useQueryClient();
  return useMutation({
    retry: false,
    mutationFn: ({
      actor,
      role,
      key,
    }: {
      actor: string;
      role: string;
      key: string;
    }) =>
      api.PATCH("/api/memberships/{actorID}", {
        params: {
          header: writeHeaders(scope.organization, key),
          path: { actorID: actor },
        },
        body: { role },
      }),
    onSettled: async () => {
      await client.invalidateQueries({ queryKey: scopeKey(scope) });
    },
  });
}
