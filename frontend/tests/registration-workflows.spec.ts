import { test, expect, type Page } from "@playwright/test";
import type {
  RegistrationConfig,
  RegistrationSubmission,
} from "@/features/registration/schemas";

const baseURL = process.env.PLAYWRIGHT_BASE_URL ?? "http://localhost:3000";
const id = "33333333-0000-4000-8000-000000000006";
const tournament = {
  id,
  name: "Registration test",
  game: "Test",
  description: "",
  organizerName: "Organizer",
  location: "Online",
  status: "REGISTRATION_OPEN",
  registrationMode: "TEAM",
  minTeamSize: 2,
  maxTeamSize: 3,
  entryFee: 0,
  currency: "THB",
  capacity: 16,
  registrationDeadline: "2099-11-01T00:00:00Z",
  startAt: "2099-11-02T00:00:00Z",
  endAt: "2099-11-03T00:00:00Z",
  startsAt: "2099-11-02T00:00:00Z",
  endsAt: "2099-11-03T00:00:00Z",
  registeredCount: 0,
};
const lobby = {
  id: "44444444-0000-4000-8000-000000000006",
  name: "Night Owls",
  inviteCode: "ABCDEF",
  status: "FORMING",
  viewerRole: "INVITEE",
  tournament,
  members: [],
};
const questions: RegistrationConfig["questions"] = [
  {
    id: "game",
    label: "Game ID",
    type: "TEXT",
    required: true,
    options: [],
    pattern: "",
    helpText: "",
    allowedTypes: [],
    maxFileBytes: 0,
  },
  {
    id: "rank",
    label: "Rank",
    type: "DROPDOWN",
    required: true,
    options: ["Gold", "Silver"],
    pattern: "",
    helpText: "",
    allowedTypes: [],
    maxFileBytes: 0,
  },
  {
    id: "proof",
    label: "Proof",
    type: "FILE",
    required: true,
    options: [],
    pattern: "",
    helpText: "",
    allowedTypes: ["application/pdf"],
    maxFileBytes: 1024,
  },
];

test.beforeEach(async ({ page, context }) => {
  await context.addCookies([
    { name: "eventory_dev_session", value: "true", url: baseURL },
  ]);
  await page.route("**/api/v1/accounts/**", async (route) => {
    const body = new URL(route.request().url()).pathname.endsWith(
      "/organizer-profile"
    )
      ? { id, organizerName: "Organizer", organizerEmail: "alex@test.gg" }
      : {
          id: "99999999-0000-4000-8000-000000000001",
          displayName: "Alex",
          handle: "alex",
          email: "alex@test.gg",
          status: "ACTIVE",
        };
    await route.fulfill({
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });
});

test("organizer creates and edits text, dropdown and file questions", async ({
  page,
}) => {
  let form = {
    version: 0,
    questions: [] as RegistrationConfig["questions"],
    consentNotice: "Consent",
  };
  let created: { registrationForm: RegistrationConfig } | undefined;
  let failSave = true;
  let formReads = 0;
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request(),
      path = new URL(request.url()).pathname;
    if (path.includes("/accounts/")) return route.fallback();
    let body: unknown = {},
      status = 200;
    if (path.endsWith("/registration-form")) {
      if (request.method() === "PUT") {
        if (failSave) {
          status = 422;
          body = { detail: "Invalid validation pattern." };
        } else {
          form = { ...request.postDataJSON(), version: form.version + 1 };
          body = form;
        }
      } else {
        formReads++;
        body = form;
      }
    } else if (request.method() === "POST" && path === "/api/v1/tournaments") {
      created = request.postDataJSON();
      form = { ...created!.registrationForm, version: 1 };
      body = tournament;
      status = 201;
    } else if (path.endsWith("/dashboard")) {
      body = {
        summary: {
          tournament,
          published: true,
          metrics: {
            acceptedEntries: 0,
            availableSpots: 16,
            confirmedParticipants: 0,
            formingEntries: 0,
            lockedEntries: 0,
            rejectedEntries: 0,
            totalEntries: 0,
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
        entries: [],
      };
    }
    await route.fulfill({
      status,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });
  await page.goto(`${baseURL}/organizer/tournaments/new`);
  await page
    .getByRole("button", { name: "Add registration form", exact: true })
    .click();
  await page.getByRole("button", { name: "Add question", exact: true }).click();
  await page
    .getByRole("button", { name: "Use this form", exact: true })
    .click();
  await expect(
    page.getByText("Enter a question.", { exact: true })
  ).toBeVisible();
  await page.getByLabel("Question label", { exact: true }).fill("Game ID");
  await page.getByLabel("Question label", { exact: true }).press("Enter");
  expect(created).toBeUndefined();
  await page.getByRole("button", { name: "Add question", exact: true }).click();
  await page.getByLabel("Question label", { exact: true }).nth(1).fill("Rank");
  await page
    .getByRole("combobox", { name: "Answer type", exact: true })
    .nth(1)
    .click();
  await page.getByRole("option", { name: "Dropdown", exact: true }).click();
  await page.getByLabel("Dropdown options", { exact: true }).fill("Gold");
  await page
    .getByRole("button", { name: "Use this form", exact: true })
    .click();
  await expect(
    page.getByText("Add at least two unique options, one per line.", {
      exact: true,
    })
  ).toBeVisible();
  await page
    .getByLabel("Dropdown options", { exact: true })
    .fill("a".repeat(2001) + "\nSilver");
  await page
    .getByRole("button", { name: "Use this form", exact: true })
    .click();
  await expect(
    page.getByText("Use at most 2000 characters per option.", { exact: true })
  ).toBeVisible();
  await page
    .getByLabel("Dropdown options", { exact: true })
    .fill("Gold\nSilver");
  await page.getByRole("button", { name: "Add question", exact: true }).click();
  await page.getByLabel("Question label", { exact: true }).nth(2).fill("Proof");
  await page
    .getByRole("combobox", { name: "Answer type", exact: true })
    .nth(2)
    .click();
  await page.getByRole("option", { name: "File upload", exact: true }).click();
  await expect(
    page.getByLabel("Maximum file size (MB)", { exact: true })
  ).toHaveValue("5");
  await page
    .getByLabel("Participant consent notice", { exact: true })
    .fill("ก".repeat(700));
  await page
    .getByRole("button", { name: "Use this form", exact: true })
    .click();
  await expect(page.getByRole("dialog")).toHaveCount(0);
  await page
    .getByLabel("Tournament Name", { exact: true })
    .fill(tournament.name);
  await page.getByLabel("Game Title", { exact: true }).fill("Test");
  for (const [label, value] of [
    ["Registration Deadline", "2099-11-01T12:00"],
    ["Tournament Start", "2099-11-02T12:00"],
    ["Tournament End", "2099-11-03T12:00"],
  ]) {
    await page.getByLabel(label, { exact: true }).fill(value);
  }
  await page
    .getByRole("button", { name: "Publish Tournament", exact: true })
    .click();
  await expect(page).toHaveURL(`${baseURL}/organizer/tournaments/${id}`);
  expect(created?.registrationForm.questions.map((q) => q.type)).toEqual([
    "TEXT",
    "DROPDOWN",
    "FILE",
  ]);
  expect(created?.registrationForm.consentNotice).toBe("ก".repeat(700));
  await page
    .getByRole("button", { name: "Edit registration form", exact: true })
    .click();
  await page
    .getByLabel("Question label", { exact: true })
    .first()
    .fill("New game ID");
  await page.getByRole("button", { name: "Save form", exact: true }).click();
  await expect(
    page.getByText("Invalid validation pattern.", { exact: true })
  ).toBeVisible();
  await expect(
    page.getByLabel("Question label", { exact: true }).first()
  ).toHaveValue("New game ID");
  failSave = false;
  const readsBeforeSave = formReads;
  await page.getByRole("button", { name: "Save form", exact: true }).click();
  await expect(page.getByText(/Version 2 · Saved/)).toBeVisible();
  expect(formReads - readsBeforeSave).toBe(1);
  await page
    .getByRole("button", { name: "Edit registration form", exact: true })
    .click();
  await expect(
    page.getByLabel("Question label", { exact: true }).first()
  ).toHaveValue("New game ID");
  await page
    .getByRole("button", { name: "Remove question 2", exact: true })
    .click();
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  expect(form.questions).toHaveLength(3);
});

async function fillRegistration(page: Page) {
  await page.getByLabel("Game ID *", { exact: true }).fill("alex_player");
  await page.getByRole("combobox", { name: "Rank *", exact: true }).click();
  await page.getByRole("option", { name: "Gold", exact: true }).click();
  await page.getByLabel("Proof *", { exact: true }).setInputFiles({
    name: "proof.pdf",
    mimeType: "application/pdf",
    buffer: Buffer.from("%PDF-1.4\nTest"),
  });
  await page.getByRole("checkbox").check();
}

test("member joins only after submission and captain creation includes answers", async ({
  page,
}) => {
  const team = structuredClone(lobby);
  let joins = 0;
  let submitted: RegistrationSubmission | undefined;
  let created:
    { name: string; registration: RegistrationSubmission } | undefined;
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request(),
      path = new URL(request.url()).pathname;
    if (path.includes("/accounts/")) return route.fallback();
    let body: unknown = {},
      status = 200;
    if (path.endsWith("/registration-form"))
      body = { version: 1, questions, consentNotice: "Original consent." };
    else if (path.endsWith("/join")) {
      joins++;
      submitted = request.postDataJSON();
      if (joins === 1) {
        status = 409;
        body = { detail: "Could not join. Try again." };
      } else {
        team.viewerRole = "MEMBER";
        body = team;
      }
    } else if (path.endsWith("/lobbies") && request.method() === "POST") {
      created = request.postDataJSON();
      team.viewerRole = "CAPTAIN";
      body = team;
      status = 201;
    } else if (path.endsWith("/lobbies/ABCDEF")) body = team;
    else if (path === `/api/v1/tournaments/${id}`)
      body = { tournament, teams: [], registeredCount: 0 };
    await route.fulfill({
      status,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });
  await page.goto(`${baseURL}/lobbies/ABCDEF`);
  await page.getByRole("button", { name: "Join team", exact: true }).click();
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  expect(joins).toBe(0);
  await page.getByRole("button", { name: "Join team", exact: true }).click();
  await page
    .getByRole("button", { name: "Submit and join", exact: true })
    .click();
  await expect(
    page.getByText("This answer is required.").first()
  ).toBeVisible();
  expect(joins).toBe(0);
  await fillRegistration(page);
  await page
    .getByRole("button", { name: "Submit and join", exact: true })
    .click();
  await expect(
    page.getByText("Could not join. Try again.", { exact: true })
  ).toBeVisible();
  await expect(page.getByLabel("Game ID *", { exact: true })).toHaveValue(
    "alex_player"
  );
  await page
    .getByRole("button", { name: "Submit and join", exact: true })
    .click();
  await expect(page.getByRole("dialog")).toHaveCount(0);
  expect(submitted?.consent).toBe(true);
  expect(submitted?.answers?.[2].file?.data).toBe(
    Buffer.from("%PDF-1.4\nTest").toString("base64")
  );
  await page.goto(`${baseURL}/tournaments/${id}/team`);
  await page.getByLabel("Team name", { exact: true }).fill("New Squad");
  await page
    .getByRole("button", { name: "Create team lobby", exact: true })
    .click();
  expect(created).toBeUndefined();
  await fillRegistration(page);
  await page
    .getByRole("button", { name: "Submit and create team", exact: true })
    .click();
  await expect(page).toHaveURL(`${baseURL}/lobbies/ABCDEF`);
  expect(created?.name).toBe("New Squad");
  expect(created?.registration).toEqual(submitted);
});

test("reconnecting cannot carry consent to a new version; reopening loads fresh questions", async ({
  page,
}) => {
  let form = { version: 1, questions: [], consentNotice: "Original consent." };
  const submissions: RegistrationSubmission[] = [];
  await page.route("**/api/v1/**", async (route) => {
    const request = route.request(),
      path = new URL(request.url()).pathname;
    if (path.includes("/accounts/")) return route.fallback();
    let body: unknown = {},
      status = 200;
    if (path.endsWith("/registration-form")) body = form;
    else if (path.endsWith("/join")) {
      const submission = request.postDataJSON();
      submissions.push(submission);
      if (submission.formVersion !== form.version) {
        status = 409;
        body = {
          detail: "registration form changed; reopen the form and try again",
        };
      } else body = { ...lobby, viewerRole: "MEMBER" };
    } else if (path.endsWith("/lobbies/ABCDEF")) body = lobby;
    await route.fulfill({
      status,
      contentType: "application/json",
      body: JSON.stringify(body),
    });
  });
  await page.goto(`${baseURL}/lobbies/ABCDEF`);
  await page.getByRole("button", { name: "Join team", exact: true }).click();
  await expect(
    page.getByText("Original consent.", { exact: true })
  ).toBeVisible();
  await page.getByRole("checkbox").check();
  form = { version: 2, questions: [], consentNotice: "Changed consent." };
  await page.evaluate(async () => {
    window.dispatchEvent(new Event("offline"));
    await new Promise((resolve) => setTimeout(resolve, 100));
    window.dispatchEvent(new Event("online"));
  });
  await page
    .getByRole("button", { name: "Submit and join", exact: true })
    .click();
  await expect(
    page.getByText("registration form changed; reopen the form and try again", {
      exact: true,
    })
  ).toBeVisible();
  expect(submissions[0].formVersion).toBe(1);
  await expect(
    page.getByText("Original consent.", { exact: true })
  ).toBeVisible();
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  await page.getByRole("button", { name: "Join team", exact: true }).click();
  await expect(
    page.getByText("Changed consent.", { exact: true })
  ).toBeVisible();
  await expect(page.getByRole("checkbox")).not.toBeChecked();
  await page
    .getByRole("button", { name: "Submit and join", exact: true })
    .click();
  expect(submissions).toHaveLength(1);
  await page.getByRole("checkbox").check();
  await page
    .getByRole("button", { name: "Submit and join", exact: true })
    .click();
  await expect(page.getByRole("dialog")).toHaveCount(0);
  expect(submissions[1].formVersion).toBe(2);
});
