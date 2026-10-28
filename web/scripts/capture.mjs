import { chromium, request } from "@playwright/test";
import { randomBytes, randomUUID } from "node:crypto";
import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";

if (!process.argv.includes("--allow-fixture-recovery")) {
  throw new Error(
    "Capture resets fixed synthetic fixture passwords. Use --allow-fixture-recovery only with the isolated loopback example.",
  );
}
const origin = "https://127.0.0.1:5173";
const evidence = fileURLToPath(new URL("../evidence/", import.meta.url));
await mkdir(evidence, { recursive: true });
const browser = await chromium.launch({
  channel: "chrome",
  headless: !process.argv.includes("--open"),
});
const context = await browser.newContext({
  ignoreHTTPSErrors: true,
  viewport: { width: 1440, height: 1000 },
});
const http = await request.newContext({
  baseURL: origin,
  ignoreHTTPSErrors: true,
  extraHTTPHeaders: { Origin: origin },
});
const page = await context.newPage();
const errors = [];
page.on("pageerror", (error) => errors.push(error.message));

async function login(name) {
  // Credentials are recovered through the real lifecycle, kept only in memory,
  // and never printed or included in screenshots/storage-state/traces.
  const email = `${name}@example.test`;
  const password = randomBytes(24).toString("base64url");
  if ((await http.post("/auth/recovery", { data: { email } })).status() !== 202)
    throw new Error("Fixture recovery unavailable");
  const messages = await (
    await http.get("http://127.0.0.1:8025/api/v1/messages?limit=100")
  ).json();
  const message = messages.messages.find(
    (item) =>
      item.Subject.includes("recover_password") &&
      item.To.some((recipient) => recipient.Address === email),
  );
  if (!message)
    throw new Error(
      "Run the isolated browser walkthrough to provision fixtures first",
    );
  const detail = await (
    await http.get(`http://127.0.0.1:8025/api/v1/message/${message.ID}`)
  ).json();
  const token = String(detail.Text).match(
    /^Token: ([A-Za-z0-9_-]{43})\s*$/m,
  )?.[1];
  if (
    !token ||
    (await http.post("/auth/reset", { data: { token, password } })).status() !==
      200
  )
    throw new Error("Fixture reset unavailable");
  await page.goto(origin);
  await page.getByLabel("Email address", { exact: true }).fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByRole("button", { name: "Sign in", exact: true }).click();
  await page
    .getByRole("heading", { name: "Make good things happen." })
    .waitFor();
}
async function screenshot(name) {
  await page.locator(".loading").waitFor({ state: "hidden" });
  await page.evaluate(() => document.fonts.ready);
  await page.screenshot({
    path: `${evidence}/${name}.png`,
    fullPage: name !== "create-request",
  });
  if (
    !(await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ))
  )
    throw new Error(`Overflow in ${name}`);
}

try {
  await page.goto(origin);
  await page.getByRole("heading", { name: "Sign in", exact: true }).waitFor();
  await screenshot("signin-desktop");
  await login("bob");
  const params = { Origin: origin, "X-Organization-ID": "northstar" };
  const existing = await (
    await context.request.get(`${origin}/api/purchases`, { headers: params })
  ).json();
  const samples = [
    ["Optical calibration kit", 680000, true],
    ["Portable field monitors", 245000, true],
    ["Cryogenic sample carriers", 189000, false],
    ["Environmental sensor array", 425000, false],
    ["Field logistics supplies", 82500, false],
  ];
  for (const [title, amount_cents, submit] of samples) {
    if (existing.data.some((record) => record.title === title)) continue;
    const response = await context.request.post(`${origin}/api/purchases`, {
      headers: { ...params, "Idempotency-Key": randomUUID() },
      data: { title, amount_cents },
    });
    if (response.status() !== 201)
      throw new Error("Synthetic purchase creation failed");
    const created = await response.json();
    if (submit) {
      const submitted = await context.request.post(
        `${origin}/api/purchases/${created.data.id}/submit`,
        { headers: { ...params, "Idempotency-Key": randomUUID() } },
      );
      if (submitted.status() !== 200)
        throw new Error("Synthetic submission failed");
    }
  }
  await page.reload();
  await page
    .getByRole("link", { name: "Optical calibration kit", exact: true })
    .waitFor();
  await page.getByRole("button", { name: /Export JSON/ }).waitFor();
  await screenshot("purchasing-desktop");
  await page.getByRole("button", { name: "New request" }).click();
  await page.getByRole("dialog").waitFor();
  await screenshot("create-request");
  await page.keyboard.press("Escape");
  await page.getByRole("dialog").waitFor({ state: "hidden" });
  if (
    !(await page
      .getByRole("button", { name: "New request" })
      .evaluate((element) => element === document.activeElement))
  )
    throw new Error("Dialog did not restore trigger focus");
  await page
    .getByRole("link", { name: "Optical calibration kit", exact: true })
    .click();
  await page
    .getByRole("heading", { name: "Optical calibration kit", exact: true })
    .waitFor();
  await screenshot("purchase-detail");
  await page.getByRole("link", { name: "Activity", exact: true }).click();
  await page
    .getByRole("heading", { name: "Every decision has a story." })
    .waitFor();
  await screenshot("activity-desktop");
  await page.getByRole("link", { name: "Purchasing", exact: true }).click();
  await page.setViewportSize({ width: 375, height: 812 });
  await screenshot("purchasing-mobile");
  await page.setViewportSize({ width: 320, height: 812 });
  await screenshot("purchasing-small-mobile");
  await page.setViewportSize({ width: 1440, height: 1000 });
  await page.getByRole("button", { name: /Sign out/ }).click();
  await page.getByRole("heading", { name: "Sign in", exact: true }).waitFor();
  await login("carol");
  await page.getByRole("link", { name: /People & roles/ }).click();
  await page
    .getByRole("heading", { name: "Organization members", exact: true })
    .waitFor();
  await screenshot("members-desktop");
  if (errors.length) throw new Error(errors.join("\n"));
  console.log(
    "Captured desktop/mobile screens; dialog focus and viewport bounds passed. No credential artifacts saved.",
  );
  if (process.argv.includes("--open")) {
    await page.goto(origin);
    console.log(
      "Review browser is open with a synthetic administrator session.",
    );
    await new Promise((resolve) => browser.once("disconnected", resolve));
  }
} finally {
  await http.dispose();
  await context.close().catch(() => undefined);
  await browser.close();
}
