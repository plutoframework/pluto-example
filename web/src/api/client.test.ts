import { expect, test, vi, afterEach } from "vitest";
import { api, authHeaders, writeHeaders } from "./client";
import { ApiError } from "./errors";

afterEach(() => vi.unstubAllGlobals());

test("transport uses same-origin cookies and stable caller-owned action keys", async () => {
  const requests: Request[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn(async (request: Request) => {
      requests.push(request);

      return Response.json({ data: { id: "purchase" } }, { status: 201 });
    }),
  );
  const header = writeHeaders("northstar", "same-logical-action");

  for (let attempt = 0; attempt < 2; attempt++) {
    await api.POST("/api/purchases", {
      params: { header },
      body: {
        title: "Optics",
        amount_cents: 2999,
      },
    });
  }

  expect(requests).toHaveLength(2);

  for (const request of requests) {
    expect(request.credentials).toBe("same-origin");
    expect(request.cache).toBe("no-store");
    expect(request.headers.get("X-Organization-ID")).toBe("northstar");
    expect(request.headers.get("Idempotency-Key")).toBe("same-logical-action");
    expect(request.headers.has("Authorization")).toBe(false);
  }
});

test("both error envelopes are sanitized and auth writes have no replay key", async () => {
  for (const body of [
    { code: "invalid_credentials" },
    {
      error: {
        code: "forbidden",
        request_id: "synthetic-request",
        message: "sensitive server detail",
      },
    },
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (request: Request) => {
        expect(request.headers.has("Idempotency-Key")).toBe(false);

        return Response.json(body, { status: 403 });
      }),
    );
    await expect(
      api.POST("/auth/signin", {
        params: { header: authHeaders() },
        body: {
          email: "synthetic@example.test",
          password: "synthetic-only",
        },
      }),
    ).rejects.toBeInstanceOf(ApiError);

    try {
      await api.GET("/auth/session");
    } catch (error) {
      expect(String(error)).not.toContain("sensitive server detail");
    }
  }
});
