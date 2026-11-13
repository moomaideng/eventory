import type { components } from "@/lib/api/schema";
import { minorToWhole, wholeToMinor } from "@/lib/money";
import type { DetailsSection } from "@/features/organizer/components/edit-tournament-dialog";
import {
  editCrowdfundingSectionSchema,
  editGateSectionSchema,
  editGeneralSectionSchema,
  editTimelineSectionSchema,
} from "@/features/organizer/schemas";
import {
  toLocalInput,
  type OrganizerSummary,
} from "@/features/organizer/utils";

type UpdateTournamentBody = components["schemas"]["UpdateTournamentRequest"];

/** Raw form state; numbers stay strings so inputs can be cleared while typing. */
export type DetailsFormValues = {
  name: string;
  game: string;
  location: string;
  description: string;
  detail: string;
  registrationMode: "SOLO" | "TEAM";
  minTeamSize: string;
  maxTeamSize: string;
  capacity: string;
  entryFee: string;
  registrationDeadline: string;
  startAt: string;
  endAt: string;
  fundingGoal: string;
  fundingDeadline: string;
};

export type DetailsField = keyof DetailsFormValues;
export type DetailsErrors = Partial<Record<DetailsField, string>>;

export const DETAILS_SECTIONS: DetailsSection[] = [
  "general",
  "registration",
  "schedule",
  "crowdfunding",
];

export const SECTION_FIELDS: Record<DetailsSection, DetailsField[]> = {
  general: ["name", "game", "location", "description", "detail"],
  registration: [
    "registrationMode",
    "minTeamSize",
    "maxTeamSize",
    "capacity",
    "entryFee",
  ],
  schedule: ["registrationDeadline", "startAt", "endAt"],
  crowdfunding: ["fundingGoal", "fundingDeadline"],
};

const EDITABLE_STATUSES: Record<DetailsSection, string[]> = {
  general: ["DRAFT", "CROWDFUNDING", "REGISTRATION", "ONGOING"],
  registration: ["DRAFT", "CROWDFUNDING"],
  schedule: ["DRAFT", "CROWDFUNDING", "REGISTRATION"],
  crowdfunding: ["DRAFT"],
};

export const SECTION_LOCK_REASON: Record<DetailsSection, string> = {
  general: "Read-only after completion",
  registration: "Locked once registration opens",
  schedule: "Locked once the tournament starts",
  crowdfunding: "Locked after draft",
};

export function isSectionEditable(section: DetailsSection, status: string) {
  return EDITABLE_STATUSES[section].includes(status);
}

export function showsCrowdfunding(summary: OrganizerSummary) {
  return (
    summary.tournament.status === "DRAFT" ||
    summary.funding.goalAmountMinor > 0
  );
}

export function initialDetailsValues(
  summary: OrganizerSummary
): DetailsFormValues {
  const { tournament, funding } = summary;
  return {
    name: tournament.name,
    game: tournament.game,
    location: tournament.location,
    description: tournament.description,
    detail: tournament.detail ?? "",
    registrationMode: tournament.registrationMode === "SOLO" ? "SOLO" : "TEAM",
    minTeamSize: String(tournament.minTeamSize || 1),
    maxTeamSize: String(tournament.maxTeamSize || 1),
    capacity: String(tournament.capacity || ""),
    entryFee: String(minorToWhole(tournament.entryFeeMinor ?? 0)),
    registrationDeadline: toLocalInput(tournament.registrationDeadline),
    startAt: toLocalInput(tournament.startAt),
    endAt: toLocalInput(tournament.endAt),
    fundingGoal:
      funding.goalAmountMinor > 0
        ? String(minorToWhole(funding.goalAmountMinor))
        : "",
    fundingDeadline: toLocalInput(funding.deadline),
  };
}

/** Fields whose value differs from the initial snapshot. Team sizes are ignored for SOLO. */
export function changedFields(
  values: DetailsFormValues,
  initial: DetailsFormValues
): DetailsField[] {
  return (Object.keys(values) as DetailsField[]).filter((field) => {
    if (
      values.registrationMode === "SOLO" &&
      (field === "minTeamSize" || field === "maxTeamSize")
    ) {
      return false;
    }
    return values[field] !== initial[field];
  });
}

function collectIssues(
  errors: DetailsErrors,
  issues: { path: PropertyKey[]; message: string }[]
) {
  for (const issue of issues) {
    const field = issue.path[0] as DetailsField | undefined;
    if (field && !errors[field]) errors[field] = issue.message;
  }
}

export type DetailsSubmission =
  | { ok: true; body: UpdateTournamentBody }
  | { ok: false; errors: DetailsErrors };

/**
 * Validates every editable section that has changes and builds a PATCH body
 * containing only changed fields the current status allows.
 */
export function buildDetailsSubmission(
  values: DetailsFormValues,
  initial: DetailsFormValues,
  status: string
): DetailsSubmission {
  const changed = new Set(changedFields(values, initial));
  const active = (section: DetailsSection) =>
    isSectionEditable(section, status) &&
    SECTION_FIELDS[section].some((field) => changed.has(field));

  const errors: DetailsErrors = {};
  const body: UpdateTournamentBody = {};

  if (active("general")) {
    const parsed = editGeneralSectionSchema.safeParse(values);
    if (!parsed.success) collectIssues(errors, parsed.error.issues);
    else {
      for (const field of SECTION_FIELDS.general) {
        if (changed.has(field)) {
          body[field as "name"] = parsed.data[field as "name"];
        }
      }
    }
  }

  if (active("registration")) {
    const solo = values.registrationMode === "SOLO";
    const parsed = editGateSectionSchema.safeParse({
      registrationMode: values.registrationMode,
      minTeamSize: solo ? 1 : values.minTeamSize,
      maxTeamSize: solo ? 1 : values.maxTeamSize,
      capacity: values.capacity,
      entryFee: values.entryFee,
    });
    if (!parsed.success) collectIssues(errors, parsed.error.issues);
    else {
      if (changed.has("registrationMode")) {
        body.registrationMode = parsed.data.registrationMode;
      }
      if (!solo && changed.has("minTeamSize")) {
        body.minTeamSize = parsed.data.minTeamSize;
      }
      if (!solo && changed.has("maxTeamSize")) {
        body.maxTeamSize = parsed.data.maxTeamSize;
      }
      if (changed.has("capacity")) body.capacity = parsed.data.capacity;
      if (changed.has("entryFee")) {
        body.entryFeeMinor = wholeToMinor(parsed.data.entryFee);
      }
    }
  }

  if (active("schedule")) {
    const dateFields = SECTION_FIELDS.schedule as (
      "registrationDeadline" | "startAt" | "endAt"
    )[];
    for (const field of dateFields) {
      if (initial[field] && !values[field]) {
        errors[field] = "This date can't be cleared once set.";
      }
    }
    const parsed = editTimelineSectionSchema.safeParse(values);
    if (!parsed.success) collectIssues(errors, parsed.error.issues);
    for (const field of dateFields) {
      if (changed.has(field) && values[field]) {
        body[field] = new Date(values[field]).toISOString();
      }
    }
  }

  if (active("crowdfunding")) {
    const parsed = editCrowdfundingSectionSchema.safeParse(values);
    if (!parsed.success) collectIssues(errors, parsed.error.issues);
    else {
      if (changed.has("fundingGoal")) {
        body.fundingGoalMinor = wholeToMinor(parsed.data.fundingGoal);
      }
      if (changed.has("fundingDeadline")) {
        body.fundingDeadline = new Date(
          parsed.data.fundingDeadline
        ).toISOString();
      }
    }
  }

  if (
    status === "DRAFT" &&
    (active("schedule") || active("crowdfunding")) &&
    values.fundingDeadline &&
    values.registrationDeadline &&
    new Date(values.fundingDeadline).getTime() >=
      new Date(values.registrationDeadline).getTime()
  ) {
    const field = active("crowdfunding")
      ? "fundingDeadline"
      : "registrationDeadline";
    errors[field] ??= "Crowdfunding must end before registration closes.";
  }

  if (Object.keys(errors).length > 0) return { ok: false, errors };
  return { ok: true, body };
}
