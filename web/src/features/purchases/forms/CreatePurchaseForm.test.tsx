import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, test, vi } from "vitest";
import { CreatePurchaseForm, parseAmount } from "./CreatePurchaseForm";

test("decimal money stays in exact integer cents and rejects ambiguous input", () => {
  expect(parseAmount("0.01")).toBe(1);
  expect(parseAmount("19.99")).toBe(1999);
  expect(parseAmount("1000000.00")).toBe(100_000_000);
  for (const input of [
    "0",
    "-1",
    "1e2",
    "1.001",
    "1000000.01",
    " 10",
    "1,000",
    "NaN",
    ".10",
    "1.",
  ])
    expect(parseAmount(input)).toBeNull();
  for (let cents = 1; cents < 100000; cents += 977) {
    const value = `${Math.floor(cents / 100)}.${String(cents % 100).padStart(2, "0")}`;
    expect(parseAmount(value)).toBe(cents);
  }
});

test("form sends a typed draft only after valid input", async () => {
  const submit = vi.fn().mockResolvedValue(undefined);
  const user = userEvent.setup();
  render(
    <CreatePurchaseForm pending={false} onSubmit={submit} onCancel={vi.fn()} />,
  );
  await user.type(
    screen.getByRole("textbox", { name: /what does your team need/i }),
    "Field optics",
  );
  await user.type(
    screen.getByRole("textbox", { name: /estimated cost/i }),
    "1.001",
  );
  await user.click(screen.getByRole("button", { name: "Create draft" }));
  expect(submit).not.toHaveBeenCalled();
  expect(screen.getByRole("alert")).toHaveTextContent("Enter a title");
  await user.clear(screen.getByRole("textbox", { name: /estimated cost/i }));
  await user.type(
    screen.getByRole("textbox", { name: /estimated cost/i }),
    "19.99",
  );
  await user.click(screen.getByRole("button", { name: "Create draft" }));
  expect(submit).toHaveBeenCalledExactlyOnceWith({
    title: "Field optics",
    amount_cents: 1999,
  });
});

test("busy form prevents duplicate submissions", async () => {
  render(<CreatePurchaseForm pending onSubmit={vi.fn()} onCancel={vi.fn()} />);
  expect(screen.getByRole("button", { name: "Saving…" })).toBeDisabled();
});
