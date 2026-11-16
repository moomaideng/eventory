import type { Page } from "@playwright/test";
import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-5.setup";

async function numberQuestion(page: Page, label?: string) {
  await page.getByRole("button", { name: "Add Question" }).first().click();
  await page.getByRole("combobox").click();
  await page.getByRole("option", { name: "Number" }).click();
  if (label) {
    await page.getByRole("button", { name: "Edit question label" }).click();
    await page.getByPlaceholder("Enter question title...").fill(label);
    await page.getByRole("button", { name: "Confirm label" }).click();
  }
  await page.getByRole("spinbutton").first().fill("0");
  await page.getByRole("spinbutton").nth(1).fill("1000");
}

const tournamentId = (page: Page) => new URL(page.url()).pathname.split("/").at(-1)!;

const apiURL = process.env.PLAYWRIGHT_API_URL ?? "http://localhost:8080";
const auth = { Authorization: "Bearer alex-dev" };

test("TC56: valid Number question saves and appears for registrants", async ({ page }) => {
  await numberQuestion(page, "Budget amount");
  await page.getByRole("button", { name: "Save Form" }).first().click();
  const id = tournamentId(page);
  await expect.poll(async () => {
    const response = await page.request.get(`${apiURL}/api/v1/tournaments/${id}/form`);
    return response.ok() && (await response.text()).includes("Budget amount");
  }).toBe(true);
  await page.reload();
  await page.getByRole("tab", { name: "Registration questions" }).click();
  await expect(page.getByText("Budget amount").first()).toBeVisible();
  await expect(page.getByRole("spinbutton").first()).toHaveValue("0");
  await expect(page.getByRole("spinbutton").nth(1)).toHaveValue("1000");

  const now = Date.now();
  const dates = {
    registrationDeadline: new Date(now + 7 * 86400000).toISOString(),
    startAt: new Date(now + 10 * 86400000).toISOString(),
    endAt: new Date(now + 11 * 86400000).toISOString(),
    capacity: 2,
  };
  const update = await page.request.patch(`${apiURL}/api/v1/tournaments/${id}`, {
    headers: auth,
    data: dates,
  });
  expect(update.ok(), await update.text()).toBe(true);
  const open = await page.request.patch(`${apiURL}/api/v1/tournaments/${id}/status`, {
    headers: auth,
    data: { status: "REGISTRATION" },
  });
  expect(open.ok(), await open.text()).toBe(true);
  await page.goto(`/tournaments/${id}`);
  await page.getByText("Register as an individual").click();
  await expect(page.getByText("Budget amount")).toBeVisible();
  await expect(page.getByRole("spinbutton")).toHaveAttribute("min", "0");
  await expect(page.getByRole("spinbutton")).toHaveAttribute("max", "1000");
});

test("TC57: empty question label blocks saving", async ({ page }) => {
  await numberQuestion(page);
  await page.getByRole("button", { name: "Save Form" }).first().click();
  await expect(page.getByText("Field #1 must have a label.")).toBeVisible();
  page.once("dialog", (dialog) => dialog.accept());
  await page.reload();
  await page.getByRole("tab", { name: "Registration questions" }).click();
  await expect(page.getByText("No custom questions added")).toBeVisible();
});

test("TC58: reversed Number range fails to save", async ({ page }) => {
  await numberQuestion(page, "Budget amount");
  await page.getByRole("spinbutton").first().fill("1000");
  await page.getByRole("spinbutton").nth(1).fill("100");
  await page.getByRole("button", { name: "Save Form" }).first().click();
  await expect(page.getByText("Saved", { exact: true })).toHaveCount(0);
  await expect(page.locator("span.text-destructive").filter({ hasText: /./ })).toBeVisible();
  const saved = await page.request.get(`${apiURL}/api/v1/tournaments/${tournamentId(page)}/form`);
  expect(saved.status()).toBe(404);
});
