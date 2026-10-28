import { Alert, Spinner } from "@heroui/react";
import type { ReactNode } from "react";

export function Notice({
  children,
  tone = "info",
}: {
  children: ReactNode;
  tone?: "info" | "error" | "success";
}) {
  return (
    <Alert
      status={
        tone === "error" ? "danger" : tone === "success" ? "success" : "accent"
      }
      className={`notice notice-${tone}`}
      role={tone === "error" ? "alert" : "status"}
    >
      <Alert.Indicator />
      <Alert.Content>{children}</Alert.Content>
    </Alert>
  );
}
export function Loading({
  label = "Connecting to your workspace…",
}: {
  label?: string;
}) {
  return (
    <div className="loading" role="status">
      <Spinner size="sm" aria-hidden="true" />
      {label}
    </div>
  );
}
export function EmptyState({
  title,
  description,
  children,
}: {
  title: string;
  description: string;
  children?: ReactNode;
}) {
  return (
    <section className="empty-state">
      <span className="empty-mark" aria-hidden="true">
        ↗
      </span>
      <h2>{title}</h2>
      <p>{description}</p>
      {children}
    </section>
  );
}
export function PageHeader({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow: string;
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <header className="page-header">
      <div>
        <span className="eyebrow">{eyebrow}</span>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      {action}
    </header>
  );
}
