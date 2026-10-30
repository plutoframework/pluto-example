import { Button, Chip, Label, ListBox, Select } from "@heroui/react";
import type { ReactNode } from "react";
import { Link, useLocation, useNavigate } from "@tanstack/react-router";
import { useWorkspace } from "@/features/workspace/public";
import { useSignOut } from "@/features/session/public";
import { OrbitMark } from "@/shared/ui/icons";
import { Notice } from "@/shared/ui/feedback";
import { errorMessage } from "@/api/errors";

export function AppShell({ children }: { children: ReactNode }) {
  const workspace = useWorkspace();
  const navigate = useNavigate();
  const pathname = useLocation({ select: (location) => location.pathname });
  const sections: Record<string, string> = {
    "/members": "People & roles",
    "/activity": "Activity",
  };

  const section = sections[pathname] ?? "Purchasing";

  const signOut = useSignOut();

  return (
    <div className="application-layout">
      <aside className="sidebar">
        <Link to="/" className="brand">
          <OrbitMark small />
          <span>
            pluto<span className="brand-caption">RESEARCH OPERATIONS</span>
          </span>
        </Link>
        <div className="workspace-label">
          <span className="workspace-monogram">
            {workspace.scope.organization === "northstar" ? "N" : "H"}
          </span>
          <div>
            <strong>
              {workspace.organization?.name ?? "Your organization"}
            </strong>
            <span>Purchasing workspace</span>
          </div>
        </div>
        <span className="nav-label">WORKSPACE</span>
        <nav aria-label="Main navigation">
          <Link
            to="/"
            activeOptions={{ exact: true }}
            activeProps={{ className: "active" }}
          >
            <span aria-hidden="true">▦</span>Purchasing
            <span className="nav-arrow" aria-hidden="true">
              ↗
            </span>
          </Link>
          <Link to="/activity" activeProps={{ className: "active" }}>
            <span aria-hidden="true">◷</span>Activity
          </Link>
          {workspace.can("memberships:manage") && (
            <Link to="/members" activeProps={{ className: "active" }}>
              <span aria-hidden="true">♧</span>People & roles
            </Link>
          )}
        </nav>
        <div className="sidebar-bottom">
          <div className="sidebar-note">
            <span className="eyebrow">BUILT FOR WHAT'S NEXT</span>
            <p>
              Less friction.
              <br />
              More possibility.
            </p>
            <OrbitMark small />
          </div>
          <span className="environment">
            <span className="signal-dot" />
            LOCAL EXAMPLE
          </span>
        </div>
      </aside>
      <div className="application-main">
        <header className="topbar">
          <span className="topbar-title">
            Organization operations<span className="topbar-divider">/</span>
            <strong>{section}</strong>
          </span>
          <div className="topbar-actions">
            <Select
              aria-label="Organization"
              className="organization-select"
              value={workspace.scope.organization}
              onChange={(value) => {
                if (typeof value !== "string") {
                  return;
                }

                void workspace.selectOrganization(value).then(() =>
                  navigate({
                    to: "/",
                    search: {},
                  }),
                );
              }}
            >
              <Select.Trigger aria-label="Organization">
                <Select.Value />
                <Select.Indicator />
              </Select.Trigger>
              <Select.Popover>
                <ListBox aria-label="Organizations">
                  <ListBox.Item id="northstar" textValue="Northstar Research">
                    <Label>Northstar Research</Label>
                    <ListBox.ItemIndicator />
                  </ListBox.Item>
                  <ListBox.Item id="harbor" textValue="Harbor Studio">
                    <Label>Harbor Studio</Label>
                    <ListBox.ItemIndicator />
                  </ListBox.Item>
                </ListBox>
              </Select.Popover>
            </Select>
            <Chip variant="soft" size="sm" className="role-label">
              {workspace.access?.role_name ?? "Access pending"}
            </Chip>
            <Button
              className="signout"
              variant="ghost"
              size="sm"
              onClick={() => {
                void signOut.mutateAsync().catch(() => undefined);
              }}
              isDisabled={signOut.isPending}
            >
              Sign out ↗
            </Button>
          </div>
        </header>
        <main id="main-content" className="main-content">
          {signOut.error && (
            <Notice tone="error">{errorMessage(signOut.error)}</Notice>
          )}
          {children}
        </main>
        <footer className="application-footer">
          <span>Thoughtful software. Forward motion.</span>
          <span>
            POWERED BY PLUTO <span className="footer-star">✧</span>
          </span>
        </footer>
      </div>
    </div>
  );
}
