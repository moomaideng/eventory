import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-4.setup";

test("TC51: valid announcement appears publicly", async ({ page }) => {
  const uniqueTitle = `Venue update ${Date.now()}`;
  await page.getByLabel("Title").fill(uniqueTitle);
  await page.getByLabel("Message").fill("The venue is Room 301.");

  await page.getByRole("button", { name: "Post announcement" }).click();
  await expect(page.getByText("Announcement posted")).toBeVisible();
  await expect(page.getByLabel("Title")).toBeEmpty();
  await expect(page.getByLabel("Message")).toBeEmpty();

  await page.getByText("Public page", { exact: true }).click();
  await page.locator('[aria-label="Tournament announcements"]')
    .getByText("View more", { exact: true }).click();
  const item = page.locator('li[id^="announcement-"]').filter({ hasText: uniqueTitle });
  await expect(item).toContainText("The venue is Room 301.");
  await expect(item).toContainText(/ · \d{1,2} /);
});

test("TC52: empty title is rejected", async ({ page }) => {
  await expect(page.locator('li[id^="announcement-"]').first()).toBeVisible();
  const before = await page.locator('li[id^="announcement-"]').count();
  await page.getByLabel("Title").fill("");
  await page.getByLabel("Message").fill("The venue is Room 301.");

  await page.getByRole("button", { name: "Post announcement" }).click();
  await expect(page.getByText("Enter a title.", { exact: true })).toBeVisible();
  await expect(page.locator('li[id^="announcement-"]')).toHaveCount(before);
});

test("TC53: title stops at 150 characters", async ({ page }) => {
  await page.getByLabel("Title").fill("A".repeat(151));
  await page.getByLabel("Message").fill("The venue is Room 301.");

  await expect(page.getByLabel("Title")).toHaveValue("A".repeat(150));
});

test("TC54: empty message is rejected", async ({ page }) => {
  await expect(page.locator('li[id^="announcement-"]').first()).toBeVisible();
  const before = await page.locator('li[id^="announcement-"]').count();
  await page.getByLabel("Title").fill("Venue update");
  await page.getByLabel("Message").fill("");

  await page.getByRole("button", { name: "Post announcement" }).click();
  await expect(page.getByText("Enter a message.", { exact: true })).toBeVisible();
  await expect(page.locator('li[id^="announcement-"]')).toHaveCount(before);
});

test("TC55: message stops at 2,000 characters", async ({ page }) => {
  await page.getByLabel("Title").fill("Venue update");
  await page.getByLabel("Message").fill("A".repeat(2001));

  await expect(page.getByLabel("Message")).toHaveValue("A".repeat(2000));
});
