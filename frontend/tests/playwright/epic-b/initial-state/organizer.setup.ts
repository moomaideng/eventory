import { test as base, expect, type BrowserContext } from "@playwright/test";
import type { Page } from "@playwright/test";

export { expect } from "@playwright/test";

export async function openOrganizer(page: Page) {
  await page.goto("/");
  await page.waitForLoadState("networkidle");
  await page.getByRole("link", { name: /Organizer workspace/ }).click();
  await expect(page).toHaveURL(/\/organizer$/, { timeout: 15_000 });
  await expect(page.getByRole("heading", { name: "My tournaments" })).toBeVisible();
}

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
      await page.getByText("Alex Events · Alex Ventures", { exact: true }).click();
      await expect
        .poll(async () =>
          (await page.context().cookies()).find(
            (cookie) => cookie.name === "eventory_dev_session"
          )?.value
        )
        .toBe("alex-dev");
      await page.getByText("Organizer workspace", { exact: true }).click();
      await expect(page.getByRole("heading", { name: "My tournaments" })).toBeVisible();
      await provide(await page.context().storageState());
      await page.close();
    },
    { scope: "worker" },
  ],
});
