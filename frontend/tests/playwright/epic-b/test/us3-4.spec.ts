import {
  test,
  expect,
} from "@/tests/playwright/epic-b/initial-state/us3-4.setup";

test("US3-4: starts on Announcements", async ({ page }) => {
  await expect(
    page.getByRole("tab", { name: "Announcements", exact: true })
  ).toHaveAttribute("aria-selected", "true");
});
