import { test as organizerTest, expect } from "./organizer.setup";
import type { APIResponse } from "@playwright/test";

export { expect };

const apiURL = process.env.PLAYWRIGHT_API_URL ?? "http://localhost:8080";
const headers = { Authorization: "Bearer alex-dev" };

async function result(response: APIResponse) {
  expect(response.ok(), await response.text()).toBe(true);
  return response.json();
}

export const test = organizerTest.extend<{
  reviewFixture: { id: string; pendingName: string };
}>({
  reviewFixture: async ({ request }, provide, testInfo) => {
    const now = Date.now();
    const capacity = testInfo.title.includes("TC61") ? 1 : 2;
    const created = await result(await request.post(`${apiURL}/api/v1/tournaments`, {
      headers,
      data: {
        name: `Playwright Review ${now}`,
        game: "Chess",
        location: "Online",
        description: "Review application test",
        registrationMode: "SOLO",
        capacity,
        registrationDeadline: new Date(now + 7 * 86400000).toISOString(),
        startAt: new Date(now + 10 * 86400000).toISOString(),
        endAt: new Date(now + 11 * 86400000).toISOString(),
      },
    }));
    const id = created.id as string;
    await result(await request.put(`${apiURL}/api/v1/tournaments/${id}/form`, {
      headers,
      data: {
        title: "Registration form",
        fields: [{ id: "rank", type: "SHORT_TEXT", label: "Rank", required: true, order: 1, config: {} }],
      },
    }));
    await result(await request.patch(`${apiURL}/api/v1/tournaments/${id}/status`, {
      headers,
      data: { status: "REGISTRATION" },
    }));
    const register = async (token: string, rank: string) => result(await request.post(
      `${apiURL}/api/v1/tournaments/${id}/solo-register`,
      { headers: { Authorization: `Bearer ${token}` }, data: { answers: [{ fieldId: "rank", value: rank }] } }
    ));
    const accepted = await register("bob-dev", "Gold");
    const pending = await register("carter-dev", "Silver");
    await result(await request.post(`${apiURL}/api/v1/tournaments/${id}/entries/${accepted.id}/review`, {
      headers,
      data: { decision: "ACCEPT" },
    }));
    await provide({ id, pendingName: pending.name });
  },
});

test.beforeEach(async ({ page, reviewFixture }) => {
  await page.goto(`/organizer/tournaments/${reviewFixture.id}`);
  await page.getByRole("tab", { name: "Registration", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Registrations" })).toBeVisible();
});
