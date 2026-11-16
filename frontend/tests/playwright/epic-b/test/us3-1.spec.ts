import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-1.setup";

test("TC62: valid inputs create a Draft", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill(`Test Cup ${Date.now()}`);
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("A local tournament.");
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page).toHaveURL(/\/organizer\/tournaments\/[0-9a-f-]+$/);
  await expect(page.getByText("Draft", { exact: true }).first()).toBeVisible();
});

test("TC63: empty name is rejected", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("");
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("A local tournament.");
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page.getByLabel("Tournament Name")).toHaveAttribute("aria-invalid", "true");
  await expect(page).toHaveURL(/\/organizer\/tournaments\/new$/);
});

test("TC64: empty game is rejected", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("Test Cup");
  await page.getByLabel("Game Title").fill("");
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("A local tournament.");
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page.getByLabel("Game Title")).toHaveAttribute("aria-invalid", "true");
  await expect(page).toHaveURL(/\/organizer\/tournaments\/new$/);
});

test("TC65: empty location is rejected", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("Test Cup");
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("");
  await page.getByLabel("Short Description").fill("A local tournament.");
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page.getByLabel("Location")).toHaveAttribute("aria-invalid", "true");
  await expect(page).toHaveURL(/\/organizer\/tournaments\/new$/);
});

test("TC66: empty description is rejected", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("Test Cup");
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("");
  await page.getByRole("button", { name: "Create Tournament" }).click();
  await expect(page.getByLabel("Short Description")).toHaveAttribute("aria-invalid", "true");
  await expect(page).toHaveURL(/\/organizer\/tournaments\/new$/);
});

test("TC67: name stops at 160 characters", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("A".repeat(161));
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("A local tournament.");
  await expect(page.getByLabel("Tournament Name")).toHaveValue("A".repeat(160));
});

test("TC68: game stops at 80 characters", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("Test Cup");
  await page.getByLabel("Game Title").fill("A".repeat(81));
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("A local tournament.");
  await expect(page.getByLabel("Game Title")).toHaveValue("A".repeat(80));
});

test("TC69: location stops at 160 characters", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("Test Cup");
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("A".repeat(161));
  await page.getByLabel("Short Description").fill("A local tournament.");
  await expect(page.getByLabel("Location")).toHaveValue("A".repeat(160));
});

test("TC70: description stops at 500 characters", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("Test Cup");
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("A".repeat(501));
  await expect(page.getByLabel("Short Description")).toHaveValue("A".repeat(500));
});

test("TC71: details stop at 10,000 characters", async ({ page }) => {
  await page.getByLabel("Tournament Name").fill("Test Cup");
  await page.getByLabel("Game Title").fill("Valorant");
  await page.getByLabel("Location").fill("Online");
  await page.getByLabel("Short Description").fill("A local tournament.");
  await page.getByLabel("Tournament Details & Rules").fill("A".repeat(10000));
  await page.getByLabel("Tournament Details & Rules").press("A");
  await expect(page.getByLabel("Tournament Details & Rules")).toHaveValue("A".repeat(10000));
});
