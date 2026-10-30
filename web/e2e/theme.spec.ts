import { test, expect } from "@playwright/test";
import { readFileSync, writeFileSync } from "node:fs";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import { parse, stringify } from "smol-toml";

test("TOML theme reloads, follows system changes and respects authored CSS", async ({
  page,
}) => {
  const config = fileURLToPath(new URL("../../pluto.toml", import.meta.url));
  const overrides = fileURLToPath(
    new URL("../src/styles/overrides.css", import.meta.url),
  );

  const original = readFileSync(config, "utf8");
  const originalCSS = readFileSync(overrides, "utf8");
  const manifest = parse(original);

  try {
    await page.emulateMedia({ colorScheme: "dark" });
    await page.goto("/");
    await expect(
      page.getByRole("heading", {
        name: "Sign in",
        exact: true,
      }),
    ).toBeVisible();

    // This test modifies only the isolated example's build-time theme. Restore
    // the original bytes, including comments and authored CSS, even on failure.
    const frontend = manifest.frontend as Record<string, unknown>;
    const theme = frontend.theme as Record<string, unknown>;
    theme.mode = "system";
    theme.dark = { accent: "#F97316" };
    theme.light = { accent: "#7C3AED" };
    writeFileSync(config, stringify(manifest));
    const root = page.locator("html");
    const accent = () =>
      root.evaluate((element) =>
        getComputedStyle(element).getPropertyValue("--accent").trim(),
      );

    await expect(root).toHaveAttribute("data-theme", "dark");
    await expect.poll(accent).toBe("#F97316");
    await page.emulateMedia({ colorScheme: "light" });
    await expect(root).toHaveAttribute("data-theme", "light");
    await expect.poll(accent).toBe("#7C3AED");
    writeFileSync(
      overrides,
      originalCSS + '\n:root[data-theme="light"] { --accent: #008866; }\n',
    );
    await expect.poll(accent).toBe("#008866");
    theme.mode = "dark";
    writeFileSync(config, stringify(manifest));
    await expect(root).toHaveAttribute("data-theme", "dark");
    await page.reload();
    await expect(root).toHaveAttribute("data-theme", "dark");
  } finally {
    writeFileSync(config, original);
    writeFileSync(overrides, originalCSS);
    execFileSync(process.execPath, [
      fileURLToPath(new URL("../scripts/theme.mjs", import.meta.url)),
    ]);
  }
});
