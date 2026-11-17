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

for (const { name, field, error } of [
  {
    name: "TC63: empty name is rejected",
    field: "Tournament Name",
    error: "Tournament name is required.",
  },
  {
    name: "TC64: empty game is rejected",
    field: "Game Title",
    error: "Game title is required.",
  },
  {
    name: "TC65: empty location is rejected",
    field: "Location",
    error: "Location is required (e.g. 'Online' or venue address).",
  },
  {
    name: "TC66: empty description is rejected",
    field: "Short Description",
    error: "Short description is required.",
  },
]) {
  test(name, async ({ page }) => {
    await page.getByLabel("Tournament Name").fill("Test Cup");
    await page.getByLabel("Game Title").fill("Valorant");
    await page.getByLabel("Location").fill("Online");
    await page.getByLabel("Short Description").fill("A local tournament.");
    await page.getByLabel(field).fill("");

    await page.getByRole("button", { name: "Create Tournament" }).click();
    await expect(page.getByText(error, { exact: true })).toBeVisible();
    await expect(page).toHaveURL(/\/organizer\/tournaments\/new$/);
  });
}

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
