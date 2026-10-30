import { Chip } from "@heroui/react";
import { statusLabels } from "../types";

export function PurchaseStatus({ status }: { status: string }) {
  const colors: Record<string, "success" | "warning" | "danger" | undefined> = {
    approved: "success",
    submitted: "warning",
    rejected: "danger",
  };

  const color = colors[status] ?? "default";

  return (
    <Chip color={color} variant="soft" size="sm" className="purchase-status">
      {statusLabels[status] ?? "Unknown status"}
    </Chip>
  );
}
