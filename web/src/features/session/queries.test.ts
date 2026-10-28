import { expect, test, vi, afterEach } from "vitest";
import { QueryClient } from "@tanstack/react-query";
import { sessionOptions } from "./queries";
afterEach(() => vi.unstubAllGlobals());

test("expired account session clears private workspace data", async () => {
  const client = new QueryClient();
  client.setQueryData(
    ["scope", "old-account", "old-session", "northstar", "purchases"],
    ["private"],
  );
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      Response.json({ code: "unauthenticated" }, { status: 401 }),
    ),
  );
  expect(await client.fetchQuery(sessionOptions())).toBeNull();
  expect(client.getQueriesData({ queryKey: ["scope"] })).toEqual([]);
  client.clear();
});

test("account rotation discards old cache before publishing the new identity", async () => {
  const client = new QueryClient();
  client.setQueryData(["session"], {
    id: "old-session",
    actor_id: "old-account",
  });
  client.setQueryData(
    ["scope", "old-account", "old-session", "northstar"],
    ["private"],
  );
  vi.stubGlobal(
    "fetch",
    vi.fn(async () =>
      Response.json({
        data: {
          id: "new-session",
          actor_id: "new-account",
          expires_at: "2030-01-01T00:00:00Z",
        },
      }),
    ),
  );
  expect(
    (await client.fetchQuery({ ...sessionOptions(), staleTime: 0 }))?.actor_id,
  ).toBe("new-account");
  expect(client.getQueriesData({ queryKey: ["scope"] })).toEqual([]);
  client.clear();
});
