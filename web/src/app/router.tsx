import {
  createRootRoute,
  createRoute,
  createRouter,
  lazyRouteComponent,
} from "@tanstack/react-router";
import type { PurchaseSearch } from "@/features/purchases/public";
import { pageSearch } from "@/shared/lib/search";
import { Application } from "./Application";

const root = createRootRoute({ component: Application });
const purchases = createRoute({
  getParentRoute: () => root,
  path: "/",
  component: lazyRouteComponent(
    () => import("@/features/purchases/public"),
    "PurchaseListPage",
  ),
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
  component: lazyRouteComponent(
    () => import("@/features/purchases/public"),
    "PurchaseDetailPage",
  ),
});
const members = createRoute({
  getParentRoute: () => root,
  path: "/members",
  component: lazyRouteComponent(
    () => import("@/features/memberships/public"),
    "MembersPage",
  ),
  validateSearch: pageSearch,
});
const activity = createRoute({
  getParentRoute: () => root,
  path: "/activity",
  component: lazyRouteComponent(
    () => import("@/features/audit/public"),
    "ActivityPage",
  ),
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
