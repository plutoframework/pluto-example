import { expect, test, vi } from "vitest";
import { ApiError } from "@/api/errors";
import type { Verdict } from "./decision";
import { createDecisionFlow, decisionTransition } from "./decision";

test("decision selection and confirmation precede all business writes", async () => {
  const write = vi.fn(async () => {});

  const flow = createDecisionFlow(write, () => "business-key");
  expect(
    await flow.send({
      type: "confirm",
      requestId: "r1",
    }),
  ).toBe(false);
  await flow.send({
    type: "choose",
    verdict: "rejected",
  });
  expect(write).not.toHaveBeenCalled();
  await flow.send({ type: "cancel" });
  expect(write).not.toHaveBeenCalled();
  expect(flow.getSnapshot().status).toBe("viewing");
  await flow.send({
    type: "choose",
    verdict: "approved",
  });
  await flow.send({
    type: "confirm",
    requestId: "r1",
  });
  expect(write).toHaveBeenCalledWith(
    "approved",
    "business-key",
    expect.any(AbortSignal),
  );
  expect(flow.getSnapshot().status).toBe("succeeded");
});

test("duplicate confirmations write once, and disposal fences late completion", async () => {
  let resolve!: () => void;
  const response = new Promise<void>((done) => {
    resolve = done;
  });

  const write = vi.fn(
    (_verdict: Verdict, _key: string, _signal: AbortSignal) => response,
  );

  const flow = createDecisionFlow(write);
  await flow.send({
    type: "choose",
    verdict: "approved",
  });
  const pending = flow.send({
    type: "confirm",
    requestId: "r1",
  });

  expect(
    await flow.send({
      type: "confirm",
      requestId: "r2",
    }),
  ).toBe(false);
  expect(write).toHaveBeenCalledOnce();
  const signal = write.mock.calls[0]?.[2];
  flow.dispose();
  expect(signal?.aborted).toBe(true);
  resolve();
  await pending;
  expect(flow.getSnapshot().status).toBe("submitting");
});

test("uncertain retries retain their business key and use new UI request IDs", async () => {
  const write = vi.fn(
    async (_verdict: Verdict, _key: string, _signal: AbortSignal) => {
      throw new ApiError(503, "internal");
    },
  );

  const newKey = vi.fn(() => "stable-business-key");
  const flow = createDecisionFlow(write, newKey);
  await flow.send({
    type: "choose",
    verdict: "approved",
  });
  await flow.send({
    type: "confirm",
    requestId: "r1",
  });
  expect(flow.getSnapshot()).toMatchObject({ status: "failed" });
  await flow.send({ type: "dismiss" });
  await flow.send({
    type: "choose",
    verdict: "approved",
  });
  await flow.send({
    type: "confirm",
    requestId: "r2",
  });
  expect(write.mock.calls.map((call) => call[1])).toEqual([
    "stable-business-key",
    "stable-business-key",
  ]);
  expect(newKey).toHaveBeenCalledOnce();
  const stale = decisionTransition(
    {
      status: "submitting",
      verdict: "approved",
      requestId: "r2",
    },
    {
      type: "resolved",
      requestId: "r1",
    },
  );

  expect(stale.accepted).toBe(false);
});
