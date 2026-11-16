import { test as organizerTest, expect } from "./organizer.setup";

export { expect };
export const test = organizerTest;

test.beforeEach(async ({ page }) => {
  await page.goto("/organizer/tournaments/new");
  await page.waitForLoadState("networkidle");
  await page.getByLabel("Tournament Name").fill(`Playwright Number Form ${Date.now()}`);
  await page.getByLabel("Game Title").fill("Chess");
  await page.getByLabel("Short Description").fill("Test form builder");
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page).toHaveURL(/\/organizer\/tournaments\/[0-9a-f-]+$/);
  await page.getByRole("tab", { name: "Registration questions" }).click();
  await expect(page.getByRole("heading", { name: "Registration Questions" })).toBeVisible();
});
