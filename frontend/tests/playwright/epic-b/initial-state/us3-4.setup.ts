import { test as organizerTest, expect, openOrganizer } from "./organizer.setup";

export { expect };
export const test = organizerTest;

test.beforeEach(async ({ page }) => {
  await openOrganizer(page);
  await page.getByRole("button", { name: "Manage SEA Valorant Challengers Open" }).click();
  await page.getByRole("tab", { name: "Announcements", exact: true }).click();
  await expect(page.getByRole("tab", { name: "Announcements", exact: true }))
    .toHaveAttribute("aria-selected", "true");
});
