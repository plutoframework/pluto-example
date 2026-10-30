import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { expect, test, vi } from "vitest";
import { PurchaseDecisionControls } from "./PurchaseDecision";
import { createDecisionFlow } from "../flows/decision";

test("confirmation, cancellation and decline use the real HeroUI controls", async () => {
  const write = vi.fn(async () => {});

  const flow = createDecisionFlow(write);
  render(<PurchaseDecisionControls flow={flow} />);
  fireEvent.click(screen.getByRole("button", { name: "Approve request" }));
  expect(
    await screen.findByRole("dialog", { name: "Approve this request?" }),
  ).toBeVisible();
  expect(write).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
  await waitFor(() =>
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
  );
  expect(write).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Decline request" }));
  await screen.findByRole("dialog", { name: "Decline this request?" });
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Confirm decline" }));
  });
  expect(write).toHaveBeenCalledOnce();
  expect(write).toHaveBeenCalledWith(
    "rejected",
    expect.any(String),
    expect.any(AbortSignal),
  );
  expect(await screen.findByText(/Decision recorded/)).toBeVisible();
  flow.dispose();
});

test("decision progress is visible before a held write completes", async () => {
  let finish!: () => void;
  const flow = createDecisionFlow(
    () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      }),
  );

  render(<PurchaseDecisionControls flow={flow} />);
  fireEvent.click(screen.getByRole("button", { name: "Approve request" }));
  await screen.findByRole("dialog");
  fireEvent.click(screen.getByRole("button", { name: "Confirm approval" }));
  expect(screen.getByRole("status")).toHaveTextContent(
    "Recording your decision",
  );
  expect(
    screen.getByRole("button", { name: "Recording decision…" }),
  ).toBeDisabled();
  await act(async () => {
    finish();
  });
  expect(await screen.findByText(/Decision recorded/)).toBeVisible();
  flow.dispose();
});
