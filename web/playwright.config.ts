import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  preserveOutput: "never",
  timeout: 30_000,
  reporter: "list",
  use: {
    baseURL: "https://127.0.0.1:5173",
    browserName: "chromium",
    channel: "chrome",
    ignoreHTTPSErrors: true,
    // Real credential tests must never persist traces or browser storage files.
    trace: "off",
    screenshot: "off",
    video: "off",
  },
});
