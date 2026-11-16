import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/playwright",
  workers: 1,
  retries: 0,
  reporter: [["list"], ["html", { open: "never" }]],
  projects: [
    {
      name: "epic-b",
      testMatch: "**/epic-b/**/*.spec.ts",
    },
  ],
  use: {
    browserName: "chromium",
    baseURL: "http://localhost:3000",
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
  },
});
