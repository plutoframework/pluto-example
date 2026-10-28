import { Chip } from "@heroui/react";
import { statusLabels } from "../types";
export function PurchaseStatus({ status }: { status: string }) {
  const color =
    status === "approved"
      ? "success"
      : status === "submitted"
        ? "warning"
        : status === "rejected"
          ? "danger"
          : "default";
  return (
    <Chip color={color} variant="soft" size="sm" className="purchase-status">
      {statusLabels[status] ?? "Unknown status"}
    </Chip>
  );
}
