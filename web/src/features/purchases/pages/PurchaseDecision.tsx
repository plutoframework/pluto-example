import { Button } from "@heroui/react";
import type { Scope } from "@/features/workspace/public";
import { usePurchaseDecision } from "../mutations";
import { PurchaseDecisionControls } from "../components/PurchaseDecision";

export function PurchaseDecision({ scope, id }: { scope: Scope; id: string }) {
  const flow = usePurchaseDecision(scope, id);
  return flow ? (
    <PurchaseDecisionControls flow={flow} />
  ) : (
    <Button isDisabled>Preparing actions…</Button>
  );
}
