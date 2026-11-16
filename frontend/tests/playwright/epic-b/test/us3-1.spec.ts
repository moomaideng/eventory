import type { Page } from "@playwright/test";
import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-1.setup";

const fields = {
  name: "#tournament-name",
  game: "#tournament-game",
  location: "#tournament-location",
  description: "#tournament-desc",
  detail: "#tournament-detail",
};

async function validInputs(page: Page) {
  await page.locator(fields.name).fill(`Test Cup ${Date.now()}`);
  await page.locator(fields.game).fill("Valorant");
  await page.locator(fields.location).fill("Online");
  await page.locator(fields.description).fill("A local tournament.");
}

async function submitInvalid(page: Page, field: string) {
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page.locator(field)).toHaveAttribute("aria-invalid", "true");
  await expect(page).toHaveURL(/\/organizer\/tournaments\/new$/);
}

test("TC62: valid inputs create a Draft", async ({ page }) => {
  await validInputs(page);
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page).toHaveURL(/\/organizer\/tournaments\/[0-9a-f-]+$/);
  await expect(page.getByText("Draft", { exact: true }).first()).toBeVisible();
});

test("TC63: empty name is rejected", async ({ page }) => {
  await validInputs(page);
  await page.locator(fields.name).fill("");
  await submitInvalid(page, fields.name);
});

test("TC64: empty game is rejected", async ({ page }) => {
  await validInputs(page);
  await page.locator(fields.game).fill("");
  await submitInvalid(page, fields.game);
});

test("TC65: empty location is rejected", async ({ page }) => {
  await validInputs(page);
  await page.locator(fields.location).fill("");
  await submitInvalid(page, fields.location);
});

test("TC66: empty description is rejected", async ({ page }) => {
  await validInputs(page);
  await page.locator(fields.description).fill("");
  await submitInvalid(page, fields.description);
});

test("TC67: name stops at 160 characters", async ({ page }) => {
  await page.locator(fields.name).fill("A".repeat(161));
  await expect(page.locator(fields.name)).toHaveValue("A".repeat(160));
});

test("TC68: game stops at 80 characters", async ({ page }) => {
  await page.locator(fields.game).fill("A".repeat(81));
  await expect(page.locator(fields.game)).toHaveValue("A".repeat(80));
});

test("TC69: location stops at 160 characters", async ({ page }) => {
  await page.locator(fields.location).fill("A".repeat(161));
  await expect(page.locator(fields.location)).toHaveValue("A".repeat(160));
});

test("TC70: description stops at 500 characters", async ({ page }) => {
  await page.locator(fields.description).fill("A".repeat(501));
  await expect(page.locator(fields.description)).toHaveValue("A".repeat(500));
});

test("TC71: details stop at 10,000 characters", async ({ page }) => {
  await page.locator(fields.detail).fill("A".repeat(10000));
  await page.locator(fields.detail).press("A");
  await expect(page.locator(fields.detail)).toHaveValue("A".repeat(10000));
});
