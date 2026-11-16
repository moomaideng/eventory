import { test as base, expect, type BrowserContext } from "@playwright/test";

export { expect } from "@playwright/test";

export const test = base.extend<
  object,
  { organizerStorageState: Awaited<ReturnType<BrowserContext["storageState"]>> }
>({
  storageState: async ({ organizerStorageState }, provide) => {
    await provide(organizerStorageState);
  },
  organizerStorageState: [
    async ({ browser }, provide) => {
      const page = await browser.newPage();
      await page.goto("http://localhost:3000/");
      await page.getByRole("button", { name: "Dev Quick Login" }).click();
      await page
        .getByText("Alex Events · Alex Ventures", { exact: true })
        .click();

      // Wait for the server action to set the cookie before saving the session.
      await expect
        .poll(async () => {
          const cookies = await page.context().cookies();
          return cookies.find(
            (cookie) => cookie.name === "eventory_dev_session"
          )?.value;
        })
        .toBe("alex-dev");

      await page.goto("http://localhost:3000/organizer");
      await expect(
        page.getByRole("heading", { name: "My tournaments", exact: true })
      ).toBeVisible();
      const state = await page.context().storageState();
      await page.close();
      await provide(state);
    },
    { scope: "worker" },
  ],
});

test.beforeEach(async ({ page }) => {
  await page.goto("/organizer");
  await page
    .getByRole("button", {
      name: "Manage SEA Valorant Challengers Open",
      exact: true,
    })
    .click();
  await page.getByRole("tab", { name: "Announcements", exact: true }).click();
  await expect(
    page.getByRole("tab", { name: "Announcements", exact: true })
  ).toHaveAttribute("aria-selected", "true");

  if (process.env.npm_lifecycle_event === "test:codegen") {
    console.info("US3-4 ready on Announcements. Click Record in Playwright Inspector.");
    const closed = page.waitForEvent("close", { timeout: 0 });
    await page.pause();
    // Recording can resume execution; keep the page open until recording ends.
    await closed;
    test.skip(true, "Interactive recording finished");
  }
});
