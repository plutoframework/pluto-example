import { test, expect, request } from "@playwright/test";
import type { APIRequestContext, Page } from "@playwright/test";
import { randomBytes } from "node:crypto";
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../../", import.meta.url));
const accounts: Record<
  string,
  {
    email: string;
    password: string;
    id: string;
  }
> = {};

const title = `Field optics ${randomBytes(4).toString("hex")}`;
let purchaseId = "";
let http: APIRequestContext;
const mailpit =
  process.env.PLUTO_EXAMPLE_E2E_MAILPIT ?? "http://127.0.0.1:8025";

test.describe.configure({ mode: "serial" });

async function confirmation(email: string, purpose: string): Promise<string> {
  const result = await http.get(`${mailpit}/api/v1/messages?limit=100`);
  const data = await result.json();
  const message = data.messages.find(
    (value: {
      To: { Address: string }[];
      Subject: string;
    }) =>
      value.Subject.includes(purpose) &&
      value.To.some((recipient) => recipient.Address === email),
  );

  if (!message) {
    throw new Error("Local verification message missing");
  }

  const detail = await (
    await http.get(`${mailpit}/api/v1/message/${message.ID}`)
  ).json();

  const match = String(detail.Text).match(/^Token: ([A-Za-z0-9_-]{43})\s*$/m);

  if (!match?.[1]) {
    throw new Error("Local verification token missing");
  }

  return match[1];
}

async function signIn(page: Page, name: string) {
  const account = accounts[name]!;
  await page.goto("/");
  await expect(
    page.getByRole("heading", {
      name: "Sign in",
      exact: true,
    }),
  ).toBeVisible();
  await page.getByLabel("Email address", { exact: true }).fill(account.email);
  await page.getByLabel("Password", { exact: true }).fill(account.password);
  await page
    .getByRole("button", {
      name: "Sign in",
      exact: true,
    })
    .click();
  await expect(
    page.getByRole("heading", { name: "Make good things happen." }),
  ).toBeVisible();
}

test.beforeAll(async () => {
  http = await request.newContext({
    baseURL: "https://127.0.0.1:5173",
    ignoreHTTPSErrors: true,
    extraHTTPHeaders: { Origin: "https://127.0.0.1:5173" },
  });

  for (const name of ["alice", "bob", "carol", "dana"]) {
    const email = `${name}@example.test`;
    const password = randomBytes(24).toString("base64url");
    expect(
      (
        await http.post("/auth/enroll", {
          data: {
            email,
            password,
          },
        })
      ).status(),
    ).toBe(202);
    // Explicit loopback-only fixture bootstrap. Re-runs recover the same four
    // synthetic accounts through the real lifecycle rather than bypassing auth.
    const verification = await confirmation(email, "verify_email").catch(
      () => null,
    );

    if (verification) {
      const verified = await http.post("/auth/verify", {
        data: { token: verification },
      });

      expect([200, 400]).toContain(verified.status());
    }

    expect(
      (await http.post("/auth/recovery", { data: { email } })).status(),
    ).toBe(202);
    expect(
      (
        await http.post("/auth/reset", {
          data: {
            token: await confirmation(email, "recover_password"),
            password,
          },
        })
      ).status(),
    ).toBe(200);
    expect(
      (
        await http.post("/auth/signin", {
          data: {
            email,
            password,
          },
        })
      ).status(),
    ).toBe(200);
    const session = await (await http.get("/auth/session")).json();
    accounts[name] = {
      email,
      password,
      id: session.data.actor_id,
    };
    expect((await http.post("/auth/logout")).status()).toBe(200);
  }

  const env: Record<string, string> = { ...process.env } as Record<
    string,
    string
  >;

  for (const line of readFileSync(`${root}/.env.example`, "utf8").split("\n")) {
    if (/^[A-Z][A-Z0-9_]*=/.test(line)) {
      const index = line.indexOf("=");
      env[line.slice(0, index)] ??= line.slice(index + 1);
    }
  }

  execFileSync(
    "go",
    [
      "run",
      "./cmd/dev-fixtures",
      ...Object.entries(accounts).flatMap(([name, account]) => [
        `--${name}-id`,
        account.id,
      ]),
    ],
    {
      cwd: root,
      env,
      stdio: "pipe",
    },
  );
});
test.afterAll(async () => {
  await http?.dispose();
});

test("member creates/submits a real request, isolates organizations and logs out", async ({
  page,
  context,
}) => {
  const errors: string[] = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await signIn(page, "alice");
  const cookies = await context.cookies();
  const sessionCookie = cookies.find(
    (cookie) => cookie.name === "__Host-pluto_session",
  );

  expect(sessionCookie?.secure).toBe(true);
  expect(sessionCookie?.httpOnly).toBe(true);
  expect(await page.evaluate(() => document.cookie)).not.toContain(
    "pluto_session",
  );
  await expect(page.getByRole("link", { name: /People & roles/ })).toHaveCount(
    0,
  );
  await page.getByRole("button", { name: "New request" }).click();
  await page
    .getByRole("textbox", { name: /what does your team need/i })
    .fill(title);
  await page.getByRole("textbox", { name: /estimated cost/i }).fill("2425.50");
  await page
    .getByRole("button", {
      name: "Create draft",
      exact: true,
    })
    .click();
  await expect(
    page.getByRole("heading", {
      name: title,
      exact: true,
    }),
  ).toBeVisible();
  purchaseId = page.url().split("/").at(-1)!;
  await expect(page.getByText("$2,425.50", { exact: true })).toBeVisible();

  if (process.env.PLUTO_EXAMPLE_E2E_AGENT_REVIEW === "true") {
    await page
      .getByRole("button", {
        name: "Review purchase",
        exact: true,
      })
      .click();
    await expect(
      page.getByText("Demo · deterministic suggestions", { exact: true }),
    ).toBeVisible();
    const reviewed = await page.request.get(`/api/purchases/${purchaseId}`, {
      headers: { "X-Organization-ID": "northstar" },
    });

    expect(reviewed.status()).toBe(200);
    expect((await reviewed.json()).data.status).toBe("draft");
  }

  await page.getByRole("button", { name: /Submit for review/ }).click();
  await expect(
    page.getByText("Awaiting review", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Approve request" }),
  ).toHaveCount(0);
  await page.locator(".organization-select [data-slot=select-trigger]").click();
  await page
    .getByRole("option", {
      name: "Harbor Studio",
      exact: true,
    })
    .click();
  await expect(
    page.getByRole("heading", { name: "Make good things happen." }),
  ).toBeVisible();
  await expect(
    page.getByRole("link", {
      name: title,
      exact: true,
    }),
  ).toHaveCount(0);
  await page.locator(".organization-select [data-slot=select-trigger]").click();
  await page
    .getByRole("option", {
      name: "Northstar Research",
      exact: true,
    })
    .click();
  await expect(
    page.getByRole("link", {
      name: title,
      exact: true,
    }),
  ).toBeVisible();
  await page
    .getByRole("searchbox", { name: "Search this page" })
    .fill("not-a-match");
  await expect(
    page.getByRole("heading", { name: "No matches on this page" }),
  ).toBeVisible();
  expect(new URL(page.url()).searchParams.get("q")).toBe("not-a-match");
  await page.getByRole("searchbox", { name: "Search this page" }).clear();
  await page.getByRole("button", { name: /Sign out/ }).click();
  await expect(
    page.getByRole("heading", {
      name: "Sign in",
      exact: true,
    }),
  ).toBeVisible();
  expect(errors).toEqual([]);
});

test("approver decides another actor's request and reads the real audit trail", async ({
  page,
}) => {
  await signIn(page, "bob");
  await page
    .getByRole("link", {
      name: title,
      exact: true,
    })
    .click();
  await page
    .getByRole("button", {
      name: "Approve request",
      exact: true,
    })
    .click();
  await page
    .getByRole("button", {
      name: "Confirm approval",
      exact: true,
    })
    .click();
  await expect(page.getByText("Approved", { exact: true })).toBeVisible();
  await page
    .getByRole("link", {
      name: "Activity",
      exact: true,
    })
    .click();
  await expect(
    page.getByRole("heading", { name: "Every decision has a story." }),
  ).toBeVisible();
  const approval = page
    .getByRole("listitem")
    .filter({ hasText: purchaseId })
    .getByRole("heading", {
      name: "purchase approved",
      exact: true,
    });

  await expect(page.locator(".loading")).toBeHidden();

  // Audit is ordered oldest first. Repeated runs retain real history, so the
  // new event can be beyond the first server page.
  for (let pageNumber = 0; pageNumber < 50; pageNumber += 1) {
    if (await approval.isVisible()) {
      break;
    }

    const firstEvent = page.locator(".activity-list li").first();
    const previous = await firstEvent.innerText();
    const next = page.getByRole("button", {
      name: "Next →",
      exact: true,
    });

    await expect(next).toBeEnabled();
    await next.click();
    await expect(firstEvent).not.toHaveText(previous);
    await expect(page.locator(".loading")).toBeHidden();
  }

  await expect(approval).toBeVisible();
});

test("administrator uses server role choices and is blocked from editing self", async ({
  page,
}) => {
  await signIn(page, "carol");
  await page.getByRole("link", { name: /People & roles/ }).click();
  await expect(
    page.getByRole("heading", {
      name: "Organization members",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(`member-${accounts.carol!.id}`)
      .locator("[data-slot=select-trigger]"),
  ).toBeDisabled();
  const alice = page.getByTestId(`member-${accounts.alice!.id}`);
  await alice.locator("[data-slot=select-trigger]").click();
  await page
    .getByRole("option", {
      name: "Approver",
      exact: true,
    })
    .click();
  const saved = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/memberships/${accounts.alice!.id}`),
  );

  await alice.getByRole("button", { name: "Update role" }).click();
  expect((await saved).status()).toBe(200);
  await expect(
    alice.getByRole("button", { name: "Update role" }),
  ).toBeDisabled();
  await expect(alice.locator("[data-slot=select-trigger]")).toContainText(
    "Approver",
  );
  await alice.locator("[data-slot=select-trigger]").click();
  await page
    .getByRole("option", {
      name: "Member",
      exact: true,
    })
    .click();
  const restored = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/memberships/${accounts.alice!.id}`),
  );

  await alice.getByRole("button", { name: "Update role" }).click();
  expect((await restored).status()).toBe(200);
  await expect(
    alice.getByRole("button", { name: "Update role" }),
  ).toBeDisabled();
});

test("mobile layout stays within the viewport and authorization still comes from Go", async ({
  page,
}) => {
  await page.setViewportSize({
    width: 375,
    height: 812,
  });
  await signIn(page, "alice");
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= window.innerWidth,
    ),
  ).toBe(true);
  await expect(page.getByRole("link", { name: /People & roles/ })).toHaveCount(
    0,
  );
  const denied = await page.request.get("/api/memberships", {
    headers: { "X-Organization-ID": "northstar" },
  });

  expect(denied.status()).toBe(403);
  const details = await page.request.get(`/api/purchases/${purchaseId}`, {
    headers: { "X-Organization-ID": "harbor" },
  });

  expect(details.status()).toBe(404);
});
