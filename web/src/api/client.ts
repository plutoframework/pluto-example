import createClient from "openapi-fetch";
import type { paths } from "./generated/paths";
import { ApiError } from "./errors";

// One transport boundary. Cookies are HttpOnly; no token is read or persisted by JS.
export const api = createClient<paths>({
  baseUrl: window.location.origin,
  fetch: (request) => globalThis.fetch(request),
  credentials: "same-origin",
  cache: "no-store",
});

api.use({
  async onResponse({ response }) {
    if (response.ok) {
      return response;
    }

    const body: unknown = await response
      .clone()
      .json()
      .catch(() => null);

    let code = "request_failed";
    let requestId = response.headers.get("X-Request-ID") ?? undefined;

    if (typeof body === "object" && body !== null) {
      const envelope = body as Record<string, unknown>;
      const detail =
        typeof envelope.error === "object" && envelope.error !== null
          ? (envelope.error as Record<string, unknown>)
          : envelope;

      if (typeof detail.code === "string") {
        code = detail.code;
      }

      if (typeof detail.request_id === "string") {
        requestId = detail.request_id;
      }
    }

    throw new ApiError(response.status, code, requestId);
  },
});

export function organizationHeaders(
  organization: string,
): Record<"X-Organization-ID", string> {
  if (!organization.trim()) {
    throw new Error("Verified organization selection is required");
  }

  return { "X-Organization-ID": organization };
}

// The caller creates this once per logical action and retains it across retries.
// Only declared idempotent business operations receive it; never auth endpoints.
export function writeHeaders(
  organization: string,
  key: string,
): Record<"X-Organization-ID" | "Idempotency-Key", string> {
  return {
    ...organizationHeaders(organization),
    "Idempotency-Key": key,
  };
}

export function authHeaders(): Record<"Origin", string> {

  // Fetch computes the real Origin in browsers; this also documents the API contract.
  return { Origin: window.location.origin };
}
