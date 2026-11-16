import { test as organizerTest, expect, openOrganizer } from "./organizer.setup";

export { expect };
export const test = organizerTest;

test.beforeEach(async ({ page }) => {
  await openOrganizer(page);
  await page.getByText("Host tournament", { exact: true }).click();
  await expect(page.getByRole("button", { name: "Create Tournament" })).toBeVisible();
});
