import { fireEvent, render, screen, cleanup } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { ProviderSignIn } from "./ProviderSignIn";

afterEach(cleanup);
const defaults = {
  available: true,
  pending: false,
  failed: false,
  redirecting: false,
  requestFailed: false,
  onContinue: vi.fn(),
};

test("disabled backend never offers provider sign-in", () => {
  render(<ProviderSignIn {...defaults} available={false} />);
  expect(screen.queryByRole("button")).toBeNull();
});
test("sign-in invokes its feature action and renders pending feedback", () => {
  const onContinue = vi.fn();
  const view = render(<ProviderSignIn {...defaults} onContinue={onContinue} />);
  fireEvent.click(
    screen.getByRole("button", {
      name: "Continue with your identity provider",
    }),
  );
  expect(onContinue).toHaveBeenCalledOnce();
  view.rerender(<ProviderSignIn {...defaults} redirecting />);
  expect(screen.getByRole("button", { name: "Redirecting…" })).toBeDisabled();
});
test("lookup failure cannot offer a guessed URL", () => {
  render(<ProviderSignIn {...defaults} failed />);
  expect(
    screen.getByText("Sign-in options are temporarily unavailable."),
  ).toBeVisible();
  expect(screen.queryByRole("button")).toBeNull();
});
