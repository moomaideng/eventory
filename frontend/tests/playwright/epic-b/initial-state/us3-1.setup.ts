import { test as organizerTest, expect } from "./organizer.setup";

export { expect };
export const test = organizerTest;

test.beforeEach(async ({ page }) => {
  await page.goto("/organizer/tournaments/new");
  await expect(page.getByRole("button", { name: "Create Tournament" })).toBeVisible();
});
