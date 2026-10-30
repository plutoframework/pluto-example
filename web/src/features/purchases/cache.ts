import type { QueryClient } from "@tanstack/react-query";
import { scopeKey } from "@/features/workspace/public";
import type { Scope } from "@/features/workspace/public";
import type { Purchase } from "./types";

/** Apply only confirmed records; preserve page boundaries and newer versions. */
export function confirmPurchase(
  client: QueryClient,
  scope: Scope,
  purchase: Purchase,
) {
  if (purchase.organization_id !== scope.organization) {
    throw new Error("Purchase response belongs to a different scope");
  }

  const prefix = scopeKey(scope);

  // Cancellation starts synchronously; background reads must not replace the response.
  void client.cancelQueries({ queryKey: [...prefix, "purchase", purchase.id] });
  void client.cancelQueries({ queryKey: [...prefix, "purchases"] });
  client.setQueryData<Purchase>(
    [...prefix, "purchase", purchase.id],
    (current) =>
      !current || current.version <= purchase.version ? purchase : current,
  );
  client.setQueriesData<{ data: Purchase[] }>(
    { queryKey: [...prefix, "purchases"] },
    (page) => {
      if (!page) {
        return page;
      }

      return {
        ...page,
        data: page.data.map((current) =>
          current.id === purchase.id && current.version <= purchase.version
            ? purchase
            : current,
        ),
      };
    },
  );
}
