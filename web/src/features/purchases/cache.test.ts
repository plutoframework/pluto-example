import { QueryClient } from "@tanstack/react-query";
import { expect, test } from "vitest";
import { confirmPurchase } from "./cache";
import type { Purchase } from "./types";
import { scopeKey } from "@/features/workspace/public";

const scope = {
  actor: "a",
  session: "s",
  organization: "one",
};

const record = {
  id: "purchase",
  organization_id: "one",
  version: 2,
  status: "approved",
} as Purchase;

test("confirmed data updates existing pages immediately without changing pagination or another tenant", () => {
  const client = new QueryClient();
  const key = [...scopeKey(scope), "purchases", ""];
  const other = [
    ...scopeKey({
      ...scope,
      organization: "two",
    }),
    "purchases",
    "",
  ];

  const page = {
    data: [
      {
        ...record,
        version: 1,
        status: "submitted",
      },
    ],
    page: {
      next_cursor: "cursor",
      has_more: true,
    },
  };

  client.setQueryData(key, page);
  client.setQueryData(other, page);
  confirmPurchase(client, scope, record);
  expect(client.getQueryData(key)).toEqual({
    ...page,
    data: [record],
  });
  expect(client.getQueryData(other)).toBe(page);
  expect(client.getQueryData([...scopeKey(scope), "purchase", record.id])).toBe(
    record,
  );
  confirmPurchase(client, scope, {
    ...record,
    version: 1,
    status: "submitted",
  });
  expect(client.getQueryData(key)).toEqual({
    ...page,
    data: [record],
  });
  client.clear();
});

test("a wrong-tenant confirmation cannot populate the cache", () => {
  const client = new QueryClient();
  expect(() =>
    confirmPurchase(client, scope, {
      ...record,
      organization_id: "two",
    }),
  ).toThrow();
  expect(client.getQueryCache().getAll()).toHaveLength(0);
  client.clear();
});
