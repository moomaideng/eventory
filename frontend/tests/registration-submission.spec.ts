import { test, expect } from "@playwright/test";

// Run against the local development server: npm run dev, then
// npx playwright test tests/registration-submission.spec.ts --workers=1
test("organizer reads original submissions and downloads files privately", async ({
  page,
  context,
}, testInfo) => {
  const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:3000";
  await context.addCookies([
    { name: "eventory_dev_session", value: "true", url: baseURL },
  ]);
  const id = "33333333-0000-4000-8000-000000000006";
  const members = ["Alex", "Bob", "Carol"].map((name, index) => ({
    id: `44444444-0000-4000-8000-00000000000${index + 1}`,
    accountId: `99999999-0000-4000-8000-00000000000${index + 1}`,
    displayName: name,
    handle: name.toLowerCase(),
    role: index ? "MEMBER" : "CAPTAIN",
    joinedAt: "2026-10-10T00:00:00Z",
  }));
  const tournament = {
    id,
    name: "Submission viewer test",
    game: "Test",
    organizerName: "Organizer",
    location: "Online",
    status: "REGISTRATION_OPEN",
    registrationMode: "TEAM",
    minTeamSize: 2,
    maxTeamSize: 3,
    entryFee: 0,
    currency: "THB",
    registrationDeadline: "2099-11-01T00:00:00Z",
    startAt: "2099-11-02T00:00:00Z",
    endAt: "2099-11-03T00:00:00Z",
    registeredCount: 1,
    capacity: 16,
  };
  const question = (id: string, label: string, type = "TEXT") => ({
    id,
    label,
    type,
    required: false,
    options: [],
    allowedTypes: [],
    pattern: "",
    helpText: "",
    maxFileBytes: 0,
  });
  const fileBytes = Buffer.from("%PDF-1.4\nPrivate registration file");
  const submission = {
    formVersion: 1,
    consentNotice: "Original consent notice.",
    questions: [
      question("game", "Original game ID"),
      question("rank", "Rank", "DROPDOWN"),
      question("proof", "Proof", "FILE"),
      question("optional", "Optional answer"),
    ],
    answers: [
      { questionId: "game", value: "alex-player" },
      { questionId: "rank", value: "Gold" },
      {
        questionId: "proof",
        value: "",
        file: { name: "proof.pdf", data: fileBytes.toString("base64") },
      },
    ],
  };
  const reads: string[] = [];
  const writes: string[] = [];
  let denyCarol = true;
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request();
    const path = new URL(request.url()).pathname;
    let body: unknown = {},
      status = 200;
    if (request.method() !== "GET") writes.push(path);
    if (path.endsWith("/accounts/me"))
      body = {
        id: members[0].accountId,
        displayName: "Organizer",
        handle: "organizer",
        email: "organizer@test.gg",
        status: "ACTIVE",
      };
    else if (path.endsWith("/organizer-profile"))
      body = { id, organizerName: "Organizer" };
    else if (path.endsWith("/dashboard"))
      body = {
        summary: {
          tournament,
          published: true,
          metrics: {
            acceptedEntries: 1,
            availableSpots: 15,
            confirmedParticipants: 3,
            formingEntries: 0,
            lockedEntries: 0,
            rejectedEntries: 0,
            totalEntries: 1,
          },
          funding: {
            goalAmount: 0,
            raisedAmount: 0,
            remainingAmount: 0,
            supporterCount: 0,
            percentage: 0,
            currency: "THB",
          },
        },
        entries: [
          {
            id,
            name: "Squad",
            status: "ACCEPTED",
            createdAt: "2026-10-10T00:00:00Z",
            members,
          },
        ],
      };
    else if (path.endsWith("/registration-form"))
      body = {
        version: 2,
        questions: [question("game", "Edited game ID")],
        consentNotice: "Edited consent notice.",
      };
    else if (path.endsWith("/registration")) {
      expect(request.headers().authorization).toBe("Bearer dev-token");
      reads.push(path);
      if (path.includes(members[1].id))
        body = {
          formVersion: 0,
          questions: [],
          answers: [],
          consentNotice: "Consent-only registration.",
        };
      else if (path.includes(members[2].id) && denyCarol) {
        status = 403;
        body = { detail: "You cannot access this registration data" };
      } else body = submission;
    }
    await route.fulfill({
      status,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });
  await page.goto(`${baseURL}/organizer/tournaments/${id}`);
  await expect(
    page.getByRole("button", { name: "View submission for Alex", exact: true })
  ).toBeVisible();
  expect(reads).toHaveLength(0); // No bulk loading private answers with the roster.
  await page.locator('[aria-labelledby="registrations"]').screenshot({
    path: testInfo.outputPath("registrations-desktop.png"),
    animations: "disabled",
  });
  const desktopActions = await page
    .getByRole("button", { name: /View submission for/ })
    .evaluateAll((buttons) =>
      buttons.map((button) => button.getBoundingClientRect().right)
    );
  expect(
    Math.max(...desktopActions) - Math.min(...desktopActions)
  ).toBeLessThan(1);
  await page
    .getByRole("button", { name: "View submission for Alex", exact: true })
    .click();
  const dialog = page.getByRole("dialog");
  await expect(
    dialog.getByText("Original game ID", { exact: true })
  ).toBeVisible();
  await expect(dialog.getByText("alex-player", { exact: true })).toBeVisible();
  await expect(dialog.getByText("Gold", { exact: true })).toBeVisible();
  await expect(dialog.getByText("Not provided", { exact: true })).toBeVisible();
  await expect(
    dialog.getByText("Original consent notice.", { exact: true })
  ).toBeVisible();
  await expect(dialog.getByText("Edited game ID", { exact: true })).toHaveCount(
    0
  );
  await dialog.screenshot({
    path: testInfo.outputPath("submission-viewer.png"),
    animations: "disabled",
  });
  const downloadEvent = page.waitForEvent("download");
  await dialog
    .getByRole("button", { name: "Download file", exact: true })
    .click();
  const download = await downloadEvent;
  expect(download.suggestedFilename()).toBe("proof.pdf");
  const stream = await download.createReadStream();
  if (!stream) throw new Error("No downloaded file");
  const chunks: Buffer[] = [];
  for await (const chunk of stream) chunks.push(Buffer.from(chunk));
  expect(Buffer.concat(chunks).equals(fileBytes)).toBe(true);
  await dialog
    .getByRole("button", { name: "Close", exact: true })
    .last()
    .click();
  await page
    .getByRole("button", { name: "View submission for Bob", exact: true })
    .click();
  await expect(
    dialog.getByText("No additional questions", { exact: true })
  ).toBeVisible();
  await expect(dialog.getByText("alex-player", { exact: true })).toHaveCount(0);
  await dialog
    .getByRole("button", { name: "Close", exact: true })
    .last()
    .click();
  await page
    .getByRole("button", { name: "View submission for Carol", exact: true })
    .click();
  await expect(
    dialog.getByText(/You cannot access this registration data/)
  ).toBeVisible();
  await expect(
    dialog.getByRole("button", { name: "Download file", exact: true })
  ).toHaveCount(0);
  denyCarol = false;
  await dialog.getByRole("button", { name: "Try again", exact: true }).click();
  await expect(
    dialog.getByText("Original game ID", { exact: true })
  ).toBeVisible();
  await page.setViewportSize({ width: 390, height: 844 });
  await expect(dialog).toBeVisible();
  expect(await dialog.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(
    true
  );
  expect(
    reads.every((path) => members.some((member) => path.includes(member.id)))
  ).toBe(true);
  expect(writes).toHaveLength(0);
  await dialog
    .getByRole("button", { name: "Close", exact: true })
    .last()
    .click();
  await expect(dialog).toHaveCount(0);
  const roster = page.locator('[aria-labelledby="registrations"]');
  expect(await roster.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(
    true
  );
  const mobileActions = await page
    .getByRole("button", { name: /View submission for/ })
    .evaluateAll((buttons) =>
      buttons.map((button) => button.getBoundingClientRect().right)
    );
  expect(Math.max(...mobileActions) - Math.min(...mobileActions)).toBeLessThan(
    1
  );
  await page
    .getByRole("button", { name: "View submission for Alex", exact: true })
    .click();
  await expect(
    dialog.getByText("Original game ID", { exact: true })
  ).toBeVisible();
  await dialog
    .getByRole("button", { name: "Close", exact: true })
    .last()
    .click();
  await expect(dialog).toHaveCount(0);
  await roster.screenshot({
    path: testInfo.outputPath("registrations-mobile.png"),
    animations: "disabled",
  });
  members[1].displayName = "A participant with a very long display name";
  await page.setViewportSize({ width: 320, height: 740 });
  await page.reload();
  await expect(
    page.getByRole("button", {
      name: `View submission for ${members[1].displayName}`,
      exact: true,
    })
  ).toBeVisible();
  expect(await roster.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(
    true
  );
});
