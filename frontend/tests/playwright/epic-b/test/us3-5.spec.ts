import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-5.setup";

const apiURL = process.env.PLAYWRIGHT_API_URL ?? "http://localhost:8080";
const auth = { Authorization: "Bearer alex-dev" };

test("TC56: valid Number question saves and appears for registrants", async ({ page }) => {
  await page.getByRole("button", { name: "Edit question label" }).click();
  await page.getByPlaceholder("Enter question title...").fill("Budget amount");
  await page.getByRole("button", { name: "Confirm label" }).click();
  await page.getByLabel("Min Value").fill("0");
  await page.getByLabel("Max Value").fill("1000");

  const unsavedChanges = page.getByText("Unsaved changes", { exact: true });
  await expect(unsavedChanges).toBeVisible();
  await page.getByRole("button", { name: "Save Form" }).first().click();
  await expect(unsavedChanges).toBeHidden();
  await page.reload();
  await page.getByRole("tab", { name: "Registration questions" }).click();
  await expect(page.getByText("Budget amount").first()).toBeVisible();
  await expect(page.getByLabel("Min Value")).toHaveValue("0");
  await expect(page.getByLabel("Max Value")).toHaveValue("1000");

  const id = new URL(page.url()).pathname.split("/").at(-1)!;
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
  await page.getByRole("button", { name: "Refresh dashboard" }).click();
  await page.getByText("Public page", { exact: true }).click();
  await page.getByText("Register as an individual").click();
  await expect(page.getByText("Budget amount")).toBeVisible();
  await expect(page.getByRole("spinbutton")).toHaveAttribute("min", "0");
  await expect(page.getByRole("spinbutton")).toHaveAttribute("max", "1000");
});

test("TC57: empty question label blocks saving", async ({ page }) => {
  await page.getByLabel("Min Value").fill("0");
  await page.getByLabel("Max Value").fill("1000");

  await page.getByRole("button", { name: "Save Form" }).first().click();
  await expect(page.getByText("Field #1 must have a label.", { exact: true })).toBeVisible();
  page.once("dialog", (dialog) => dialog.accept());
  await page.reload();
  await page.getByRole("tab", { name: "Registration questions" }).click();
  await expect(page.getByText("No custom questions added")).toBeVisible();
});

test("TC58: reversed Number range fails to save", async ({ page }) => {
  await page.getByRole("button", { name: "Edit question label" }).click();
  await page.getByPlaceholder("Enter question title...").fill("Budget amount");
  await page.getByRole("button", { name: "Confirm label" }).click();
  await page.getByLabel("Min Value").fill("1000");
  await page.getByLabel("Max Value").fill("100");

  await page.getByRole("button", { name: "Save Form" }).first().click();
  await expect(page.getByText("Failed to save registration form.", { exact: true })).toBeVisible();
  page.once("dialog", (dialog) => dialog.accept());
  await page.reload();
  await page.getByRole("tab", { name: "Registration questions" }).click();
  await expect(page.getByText("No custom questions added")).toBeVisible();
});
