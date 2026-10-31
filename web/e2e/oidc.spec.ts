import { expect, test } from "@playwright/test";

test.use({
  baseURL: process.env.PLUTO_EXAMPLE_WEB_TEST_URL ?? "https://127.0.0.1:5173",
});

// Browser navigation fixture only. The Go integration suite verifies actual
// signed tokens, encrypted persistence and account/session issuance separately.
test("provider sign-in posts trusted Origin and acknowledges before navigation", async ({
  page,
}) => {
  await page.route("**/auth/session", (route) =>
    route.fulfill({
      status: 401,
      contentType: "application/json",
      body: JSON.stringify({ error: { code: "unauthenticated" } }),
    }),
  );
  await page.route("**/auth/providers", (route) =>
    route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({ data: ["oidc"] }),
    }),
  );
  let release: (() => void) | undefined;
  const ready = new Promise<void>((resolve) => {
    release = resolve;
  });

  let reached: (() => void) | undefined;
  const submitted = new Promise<void>((resolve) => {
    reached = resolve;
  });

  await page.route("**/auth/oidc/begin", async (route) => {
    expect(route.request().method()).toBe("POST");
    const origin = new URL(page.url()).origin;
    expect(route.request().headers().origin).toBe(origin);
    reached?.();
    await ready;
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify({
        data: { authorization_url: origin + "/oidc-test-provider" },
      }),
    });
  });
  await page.route("**/oidc-test-provider", (route) =>
    route.fulfill({
      contentType: "text/html",
      body: "<h1>Local provider navigation fixture</h1>",
    }),
  );
  await page.goto("/");
  const click = page
    .getByRole("button", { name: "Continue with your identity provider" })
    .click();

  await submitted;
  await expect(
    page.getByRole("button", { name: "Redirecting…" }),
  ).toBeDisabled();
  release?.();
  await click;
  await expect(
    page.getByRole("heading", { name: "Local provider navigation fixture" }),
  ).toBeVisible();
});
