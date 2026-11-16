import { test, expect } from "@/tests/playwright/epic-b/initial-state/us3-4.setup";
import type { Page } from "@playwright/test";

const title = "#announcement-title";
const message = "#announcement-body";

const announcementCount = (page: Page) =>
  page.locator('li[id^="announcement-"]').count();

test("TC51: valid announcement appears publicly", async ({ page }) => {
  const tournamentId = new URL(page.url()).pathname.split("/").at(-1);
  const uniqueTitle = `Venue update ${Date.now()}`;
  await page.locator(title).fill(uniqueTitle);
  await page.locator(message).fill("The venue is Room 301.");
  await page.getByRole("button", { name: "Post announcement" }).click();
  await expect(page.getByText("Announcement posted")).toBeVisible();
  await expect(page.locator(title)).toBeEmpty();
  await expect(page.locator(message)).toBeEmpty();
  await page.goto(`/tournaments/${tournamentId}/announcements`);
  const item = page.locator('li[id^="announcement-"]').filter({ hasText: uniqueTitle });
  await expect(item).toContainText("The venue is Room 301.");
  await expect(item).toContainText(/ · \d{1,2} /);
});

test("TC52: empty title is rejected", async ({ page }) => {
  const before = await announcementCount(page);
  await page.locator(message).fill("The venue is Room 301.");
  await page.getByRole("button", { name: "Post announcement" }).click();
  await expect(page.locator(title)).toHaveAttribute("aria-invalid", "true");
  expect(await announcementCount(page)).toBe(before);
});

test("TC53: title stops at 150 characters", async ({ page }) => {
  await page.locator(title).fill("A".repeat(151));
  await expect(page.locator(title)).toHaveValue("A".repeat(150));
});

test("TC54: empty message is rejected", async ({ page }) => {
  const before = await announcementCount(page);
  await page.locator(title).fill("Venue update");
  await page.getByRole("button", { name: "Post announcement" }).click();
  await expect(page.locator(message)).toHaveAttribute("aria-invalid", "true");
  expect(await announcementCount(page)).toBe(before);
});

test("TC55: message stops at 2,000 characters", async ({ page }) => {
  await page.locator(message).fill("A".repeat(2001));
  await expect(page.locator(message)).toHaveValue("A".repeat(2000));
});
