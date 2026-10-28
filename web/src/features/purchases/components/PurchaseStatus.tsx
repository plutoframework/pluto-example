import { statusLabels } from "../types";
export function PurchaseStatus({ status }: { status: string }) {
  return (
    <span className={`status status-${status}`}>
      <span />
      {statusLabels[status] ?? "Unknown status"}
    </span>
  );
}
