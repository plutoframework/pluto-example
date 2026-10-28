import {
  createRootRoute,
  createRoute,
  createRouter,
} from "@tanstack/react-router";
import {
  PurchaseDetailPage,
  PurchaseListPage,
} from "@/features/purchases/public";
import type { PurchaseSearch } from "@/features/purchases/public";
import { MembersPage } from "@/features/memberships/public";
import { ActivityPage } from "@/features/audit/public";
import { pageSearch } from "@/shared/lib/search";
import { Application } from "./Application";

const root = createRootRoute({ component: Application });
const purchases = createRoute({
  getParentRoute: () => root,
  path: "/",
  component: PurchaseListPage,
  validateSearch: (raw: Record<string, unknown>): PurchaseSearch => ({
    ...pageSearch(raw),
    q: typeof raw.q === "string" ? raw.q.slice(0, 160) : undefined,
    status: ["draft", "submitted", "approved", "rejected"].includes(
      String(raw.status),
    )
      ? (raw.status as PurchaseSearch["status"])
      : undefined,
  }),
});
const detail = createRoute({
  getParentRoute: () => root,
  path: "/purchases/$purchaseId",
  component: PurchaseDetailPage,
});
const members = createRoute({
  getParentRoute: () => root,
  path: "/members",
  component: MembersPage,
  validateSearch: pageSearch,
});
const activity = createRoute({
  getParentRoute: () => root,
  path: "/activity",
  component: ActivityPage,
  validateSearch: pageSearch,
});
export const router = createRouter({
  routeTree: root.addChildren([purchases, detail, members, activity]),
  defaultPreload: "intent",
  defaultNotFoundComponent: () => (
    <main className="empty-state">
      <h1>Outside this orbit</h1>
      <p>
        This page does not exist. <a href="/">Return to purchasing.</a>
      </p>
    </main>
  ),
});
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
