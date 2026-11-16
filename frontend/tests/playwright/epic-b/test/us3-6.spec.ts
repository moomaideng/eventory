import type { Page } from "@playwright/test";
import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-6.setup";

const row = (page: Page, name: string) => page.getByRole("row").filter({ hasText: name });
const acceptedCount = (page: Page) => page.locator('section[aria-labelledby="registrations"] dl > div')
  .filter({ has: page.locator("dt", { hasText: "Accepted" }) }).locator("dd");

test("TC59: view answers and accept an application", async ({ page, reviewFixture }) => {
  const application = row(page, reviewFixture.pendingName);
  const before = Number(await acceptedCount(page).innerText());
  await application.getByRole("button", { name: /View answers for/ }).first().click();
  const answers = page.getByRole("dialog", { name: "Registration Answers" });
  await expect(answers).toContainText("Rank");
  await expect(answers).toContainText("Silver");
  await answers.getByRole("button", { name: "Close" }).last().click();
  await application.getByRole("button", { name: `Accept ${reviewFixture.pendingName}` }).click();
  await page.getByRole("alertdialog", { name: "Accept this application?" })
    .getByRole("button", { name: "Accept", exact: true }).click();
  await expect(application).toContainText(/accepted/i);
  await expect(acceptedCount(page)).toHaveText(String(before + 1));
});

test("TC60: reject an application", async ({ page, reviewFixture }) => {
  const application = row(page, reviewFixture.pendingName);
  const before = await acceptedCount(page).innerText();
  await application.getByRole("button", { name: `Reject ${reviewFixture.pendingName}` }).click();
  await page.getByRole("dialog", { name: "Reject this application?" })
    .getByRole("button", { name: "Reject", exact: true }).click();
  await expect(application).toContainText(/rejected/i);
  await expect(application.getByRole("button", { name: /^(Accept|Reject) / })).toHaveCount(0);
  await expect(acceptedCount(page)).toHaveText(before);
});

test("TC61: accepting at capacity is blocked", async ({ page, reviewFixture }) => {
  const application = row(page, reviewFixture.pendingName);
  const before = await acceptedCount(page).innerText();
  await application.getByRole("button", { name: `Accept ${reviewFixture.pendingName}` }).click();
  const dialog = page.getByRole("alertdialog", { name: "Accept this application?" });
  await dialog.getByRole("button", { name: "Accept", exact: true }).click();
  await expect(dialog.getByText("The application was not accepted")).toBeVisible();
  await dialog.getByRole("button", { name: "Cancel" }).click();
  await expect(application).toContainText(/waiting for approval/i);
  await expect(acceptedCount(page)).toHaveText(before);
});
