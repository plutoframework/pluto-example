import type { components } from "@/api/generated/components";

export type Purchase = components["schemas"]["Purchase"];

export type CreatePurchase = components["schemas"]["CreatePurchaseRequest"];

export type PurchaseStatus = "draft" | "submitted" | "approved" | "rejected";

export type PurchaseSearch = {
  q?: string;
  status?: PurchaseStatus;
  cursor?: string;
  trail?: string[];
};

export const statusLabels: Record<string, string> = {
  draft: "Draft",
  submitted: "Awaiting review",
  approved: "Approved",
  rejected: "Declined",
};
