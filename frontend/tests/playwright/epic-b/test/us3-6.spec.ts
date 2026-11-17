import type { Page } from "@playwright/test";
import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-6.setup";

const acceptedCount = (page: Page) => page.locator('section[aria-labelledby="registrations"] dl > div')
  .filter({ has: page.locator("dt", { hasText: "Accepted" }) }).locator("dd");

test("TC59: accept an application", async ({ page, reviewFixture }) => {
  const application = page.getByRole("row").filter({ hasText: reviewFixture.pendingName });
  await expect(acceptedCount(page)).toHaveText("1");

  await application.getByRole("button", { name: `Accept ${reviewFixture.pendingName}` }).click();
  await page.getByRole("alertdialog", { name: "Accept this application?" })
    .getByRole("button", { name: "Accept", exact: true }).click();
  await expect(application).toContainText("accepted");
  await expect(acceptedCount(page)).toHaveText("2");
});

test("TC60: reject an application", async ({ page, reviewFixture }) => {
  const application = page.getByRole("row").filter({ hasText: reviewFixture.pendingName });
  await expect(acceptedCount(page)).toHaveText("1");

  await application.getByRole("button", { name: `Reject ${reviewFixture.pendingName}` }).click();
  await page.getByRole("dialog", { name: "Reject this application?" })
    .getByRole("button", { name: "Reject", exact: true }).click();
  await expect(application).toContainText("rejected");
  await expect(application.getByRole("button", { name: `Accept ${reviewFixture.pendingName}` })).toHaveCount(0);
  await expect(application.getByRole("button", { name: `Reject ${reviewFixture.pendingName}` })).toHaveCount(0);
  await expect(acceptedCount(page)).toHaveText("1");
});

test("TC61: accepting at capacity is blocked", async ({ page, reviewFixture }) => {
  const application = page.getByRole("row").filter({ hasText: reviewFixture.pendingName });
  await expect(acceptedCount(page)).toHaveText("2");

  await application.getByRole("button", { name: `Accept ${reviewFixture.pendingName}` }).click();
  const dialog = page.getByRole("alertdialog", { name: "Accept this application?" });
  await dialog.getByRole("button", { name: "Accept", exact: true }).click();
  await expect(dialog.getByText("The tournament has reached its capacity", { exact: true })).toBeVisible();
  await dialog.getByRole("button", { name: "Cancel" }).click();
  await expect(application).toContainText("Waiting for approval");
  await expect(acceptedCount(page)).toHaveText("2");
});
