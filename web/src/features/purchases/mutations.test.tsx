import { StrictMode } from "react";
import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { afterEach, expect, test, vi } from "vitest";
import { usePurchaseDecision } from "./mutations";
import type { Scope } from "@/features/workspace/public";

afterEach(() => vi.unstubAllGlobals());

test.each(["actor", "session", "organization", "id"] as const)(
  "%s replacement aborts and fences the old purchase flow under StrictMode",
  async (field) => {
    const requests: Request[] = [];
    const responses: ((response: Response) => void)[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn((request: Request) => {
        requests.push(request);

        return new Promise<Response>((resolve) => {
          responses.push(resolve);
        });
      }),
    );
    const client = new QueryClient({
      defaultOptions: { mutations: { retry: false } },
    });

    const scope: Scope = {
      actor: "actor",
      session: "session",
      organization: "northstar",
    };

    const { result, rerender, unmount } = renderHook(
      ({ scope, id }) => usePurchaseDecision(scope, id),
      {
        initialProps: {
          scope,
          id: "purchase-1",
        },
        wrapper: ({ children }: { children: ReactNode }) => (
          <StrictMode>
            <QueryClientProvider client={client}>
              {children}
            </QueryClientProvider>
          </StrictMode>
        ),
      },
    );

    const previous = result.current!;
    expect(requests).toHaveLength(0);
    await act(async () => {
      await previous.send({
        type: "choose",
        verdict: "approved",
      });
    });
    let pending!: Promise<boolean>;
    act(() => {
      pending = previous.send({
        type: "confirm",
        requestId: "old-ui-request",
      });
    });
    await waitFor(() => expect(requests).toHaveLength(1));
    expect(requests[0]!.headers.get("X-Organization-ID")).toBe("northstar");
    expect(requests[0]!.headers.get("Idempotency-Key")).not.toBe(
      "old-ui-request",
    );
    expect(requests[0]!.headers.get("Idempotency-Key")).toBeTruthy();
    expect(JSON.parse(await requests[0]!.clone().text())).toEqual({
      decision: "approved",
    });

    rerender({
      scope:
        field === "id"
          ? scope
          : {
            ...scope,
            [field]: "replacement",
          },
      id: field === "id" ? "purchase-2" : "purchase-1",
    });
    await waitFor(() => expect(result.current).not.toBe(previous));
    expect(requests[0]!.signal.aborted).toBe(true);
    expect(result.current!.getSnapshot().status).toBe("viewing");
    expect(
      await previous.send({
        type: "resolved",
        requestId: "old-ui-request",
      }),
    ).toBe(false);
    await act(async () => {
      responses[0]!(
        Response.json({
          data: {
            id: "purchase-1",
            status: "approved",
          },
        }),
      );
      await pending;
    });
    expect(result.current!.getSnapshot().status).toBe("viewing");
    expect(previous.getSnapshot().status).toBe("submitting");
    unmount();
    expect(
      await result.current!.send({
        type: "choose",
        verdict: "approved",
      }),
    ).toBe(false);
    client.clear();
  },
);

test("confirmed writes complete before a held background refresh and update visible cache", async () => {
  const { usePurchaseCommand } = await import("./mutations");
  const { scopeKey } = await import("@/features/workspace/public");
  const scope = {
    actor: "a",
    session: "s",
    organization: "one",
  };

  const client = new QueryClient();
  const record = {
    id: "purchase",
    organization_id: "one",
    status: "submitted",
    version: 2,
  };

  let respond!: (value: Response) => void;
  vi.stubGlobal(
    "fetch",
    vi.fn(
      () =>
        new Promise<Response>((resolve) => {
          respond = resolve;
        }),
    ),
  );
  let refresh!: () => void;
  vi.spyOn(client, "invalidateQueries").mockImplementation(
    () =>
      new Promise<void>((resolve) => {
        refresh = resolve;
      }),
  );
  const key = [...scopeKey(scope), "purchases", ""];
  client.setQueryData(key, {
    data: [
      {
        ...record,
        status: "draft",
        version: 1,
      },
    ],
  });
  const { result, unmount } = renderHook(() => usePurchaseCommand(scope), {
    wrapper: ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    ),
  });

  let running!: Promise<unknown>;
  act(() => {
    running = result.current.mutateAsync({
      action: "submit",
      id: "purchase",
      key: "business-key",
    });
  });
  expect(result.current.isPending).toBe(true);
  await waitFor(() => expect(respond).toBeDefined());
  await act(async () => {
    respond(Response.json({ data: record }));
    await running;
  });
  expect(result.current.isPending).toBe(false);
  expect(client.getQueryData(key)).toEqual({ data: [record] });
  refresh();
  unmount();
  client.clear();
});
