import type { components } from "@/lib/api/schema";
import type { DetailsField } from "@/features/organizer/tournament-details-form";
import { formatRelative, type OrganizerSummary } from "@/features/organizer/utils";
import { formatMoney, formatTournamentDate } from "@/features/tournaments/utils";

// Mirrors the backend transition checks in TournamentUseCase.OverrideWithForce,
// registrationPlanProblem, validateBracketReadiness and validateCompletionReadiness.

export type LifecycleStatus =
  | "DRAFT"
  | "CROWDFUNDING"
  | "REGISTRATION"
  | "ONGOING"
  | "COMPLETED";

export type WorkspaceTab = "registration" | "form" | "sponsorship" | "bracket";

export type RequirementFix =
  | { kind: "field"; field: DetailsField }
  | { kind: "tab"; tab: WorkspaceTab };

/** met/unmet are verified; pending is loading; unknown could not be verified; info never blocks. */
export type RequirementState = "met" | "unmet" | "pending" | "unknown" | "info";

export type Requirement = {
  id: string;
  label: string;
  detail?: string;
  state: RequirementState;
  fix?: RequirementFix;
};

type Match = components["schemas"]["MatchResponse"];
type Tier = components["schemas"]["SponsorshipTierResponse"];

/** Remote data the checks depend on: undefined while loading, null when unavailable. */
export type Remote<T> = T[] | undefined | null;

const fixField = (field: DetailsField): RequirementFix => ({
  kind: "field",
  field,
});

function check(ok: boolean): RequirementState {
  return ok ? "met" : "unmet";
}

function time(value?: string | null) {
  if (!value) return null;
  const parsed = new Date(value).getTime();
  return Number.isNaN(parsed) ? null : parsed;
}

function when(value: string, now: number) {
  return `${formatRelative(value, now)} · ${formatTournamentDate(value)}`;
}

function plural(count: number, noun: string) {
  return `${count} ${noun}${count === 1 ? "" : "s"}`;
}

export function blockingCount(items: Requirement[]) {
  return items.filter((item) => item.state === "unmet").length;
}

export function isChecking(items: Requirement[]) {
  return items.some((item) => item.state === "pending");
}

export function crowdfundingRequirements(
  summary: OrganizerSummary,
  tiers: Remote<Tier>,
  now: number
): Requirement[] {
  const { tournament, funding } = summary;
  const fundingDeadline = time(funding.deadline);
  const registrationDeadline = time(tournament.registrationDeadline);
  const sellableTiers = tiers?.filter((t) => t.priceMinor > 0 && t.slots >= 1).length;

  const items: Requirement[] = [
    {
      id: "funding-goal",
      label: "Funding goal set",
      detail:
        funding.goalAmountMinor > 0
          ? formatMoney(funding.goalAmountMinor, funding.currency)
          : "No goal set yet",
      state: check(funding.goalAmountMinor > 0),
      fix: fixField("fundingGoal"),
    },
    {
      id: "funding-deadline",
      label: "Funding deadline in the future",
      detail: funding.deadline ? when(funding.deadline, now) : "Not set",
      state: check(fundingDeadline !== null && fundingDeadline > now),
      fix: fixField("fundingDeadline"),
    },
    {
      id: "tiers",
      label: "At least one sponsorship tier on sale",
      detail:
        tiers === undefined
          ? "Checking tiers…"
          : tiers === null
            ? "Couldn't load tiers; checked when you continue"
            : sellableTiers
              ? `${plural(sellableTiers, "tier")} with a price and open slots`
              : "No tier with a price and at least one slot",
      state:
        tiers === undefined
          ? "pending"
          : tiers === null
            ? "unknown"
            : check(Boolean(sellableTiers)),
      fix: { kind: "tab", tab: "sponsorship" },
    },
  ];

  if (registrationDeadline !== null && fundingDeadline !== null) {
    items.push({
      id: "funding-before-registration",
      label: "Funding closes before registration",
      detail: `Registration closes ${formatTournamentDate(tournament.registrationDeadline)}`,
      state: check(fundingDeadline < registrationDeadline),
      fix: fixField("fundingDeadline"),
    });
  }
  return items;
}

function registrationRulesProblem(
  t: OrganizerSummary["tournament"]
): { message: string; field: DetailsField } | null {
  if (t.capacity < 1) {
    return { message: "Capacity must be at least 1", field: "capacity" };
  }
  if (t.entryFeeMinor < 0) {
    return { message: "Entry fee can't be negative", field: "entryFee" };
  }
  if (t.registrationMode === "SOLO") {
    return t.minTeamSize === 1 && t.maxTeamSize === 1
      ? null
      : {
          message: "Solo events need a team size of 1",
          field: "registrationMode",
        };
  }
  if (t.registrationMode === "TEAM") {
    return t.minTeamSize >= 1 && t.maxTeamSize >= t.minTeamSize
      ? null
      : {
          message: "Set a valid team size range",
          field: t.minTeamSize >= 1 ? "maxTeamSize" : "minTeamSize",
        };
  }
  return { message: "Choose a registration mode", field: "registrationMode" };
}

function scheduleProblem(
  t: OrganizerSummary["tournament"],
  now: number
): { message: string; field: DetailsField } | null {
  const deadline = time(t.registrationDeadline);
  const startAt = time(t.startAt);
  const endAt = time(t.endAt);
  if (deadline === null) {
    return {
      message: "Registration deadline not set",
      field: "registrationDeadline",
    };
  }
  if (deadline <= now) {
    return {
      message: "Registration deadline has passed",
      field: "registrationDeadline",
    };
  }
  if (startAt === null) {
    return { message: "Event start not set", field: "startAt" };
  }
  if (endAt === null) {
    return { message: "Event end not set", field: "endAt" };
  }
  if (deadline >= startAt) {
    return {
      message: "Registration must close before the event starts",
      field: "registrationDeadline",
    };
  }
  if (startAt >= endAt) {
    return { message: "Event must end after it starts", field: "endAt" };
  }
  return null;
}

export function registrationRequirements(
  summary: OrganizerSummary,
  now: number,
  force: boolean
): Requirement[] {
  const { tournament: t, funding } = summary;
  const missingGeneral = [
    !t.name.trim() && "name",
    !t.game.trim() && "game",
    !t.location.trim() && "location",
  ].filter(Boolean);
  const rulesProblem = registrationRulesProblem(t);
  const schedule = scheduleProblem(t, now);

  const items: Requirement[] = [
    {
      id: "general",
      label: "Name, game and location",
      detail: missingGeneral.length
        ? `Missing ${missingGeneral.join(", ")}`
        : `${t.game} · ${t.location}`,
      state: check(missingGeneral.length === 0),
      fix: fixField(
        !t.name.trim() ? "name" : !t.game.trim() ? "game" : "location"
      ),
    },
    {
      id: "rules",
      label: "Registration rules",
      detail:
        rulesProblem?.message ??
        `${plural(t.capacity, t.registrationMode === "SOLO" ? "player" : "team")} · ${formatMoney(t.entryFeeMinor, t.currency, true)}`,
      state: check(rulesProblem === null),
      fix: fixField(rulesProblem?.field ?? "capacity"),
    },
    {
      id: "schedule",
      label: "Schedule: registration closes, then the event starts and ends",
      detail:
        schedule?.message ??
        `Registration closes ${when(t.registrationDeadline as string, now)}`,
      state: check(schedule === null),
      fix: fixField(schedule?.field ?? "registrationDeadline"),
    },
  ];

  if (t.status === "CROWDFUNDING") {
    const fundingDeadline = time(funding.deadline);
    const registrationDeadline = time(t.registrationDeadline);
    if (fundingDeadline !== null) {
      items.push({
        id: "funding-before-registration",
        label: "Funding closes before registration",
        detail: `Funding closes ${formatTournamentDate(funding.deadline)}`,
        state: check(
          registrationDeadline !== null && fundingDeadline < registrationDeadline
        ),
        fix: fixField("registrationDeadline"),
      });
    }
    const reached = funding.raisedAmountMinor >= funding.goalAmountMinor;
    items.push({
      id: "funding-goal-reached",
      label: "Funding goal reached",
      detail: reached
        ? `${formatMoney(funding.raisedAmountMinor, funding.currency)} raised`
        : `${formatMoney(funding.raisedAmountMinor, funding.currency)} of ${formatMoney(funding.goalAmountMinor, funding.currency)}${force ? " · opening early" : ""}`,
      state: check(reached || force),
    });
  }

  items.push({
    id: "questions",
    label: "Registration questions (optional)",
    detail: "Ask players for extra details; locked once registration opens",
    state: "info",
    fix: { kind: "tab", tab: "form" },
  });
  return items;
}

function bracketUnavailable(id: string, label: string, matches: undefined | null): Requirement {
  return {
    id,
    label,
    detail:
      matches === undefined
        ? "Checking bracket…"
        : "Couldn't load the bracket; checked when you continue",
    state: matches === undefined ? "pending" : "unknown",
    fix: { kind: "tab", tab: "bracket" },
  };
}

export function startRequirements(
  summary: OrganizerSummary,
  matches: Remote<Match>,
  now: number
): Requirement[] {
  const { tournament: t, metrics: m } = summary;
  const deadline = time(t.registrationDeadline);
  const paymentPending = Math.max(
    m.totalEntries -
      m.formingEntries -
      m.lockedEntries -
      m.acceptedEntries -
      m.rejectedEntries,
    0
  );
  const unresolved = m.formingEntries + m.lockedEntries + paymentPending;
  const accepted = m.acceptedEntries;

  const items: Requirement[] = [
    {
      id: "deadline-passed",
      label: "Registration deadline passed",
      detail: t.registrationDeadline
        ? `${deadline !== null && deadline <= now ? "Closed" : "Closes"} ${when(t.registrationDeadline, now)}`
        : "Not set",
      state: check(deadline !== null && deadline <= now),
    },
    {
      id: "unresolved",
      label: "No unresolved applications",
      detail: unresolved
        ? `${plural(unresolved, "application")} still forming, awaiting payment or review`
        : "Every application is accepted or rejected",
      state: check(unresolved === 0),
      fix: { kind: "tab", tab: "registration" },
    },
    {
      id: "accepted",
      label: "At least 2 accepted teams",
      detail: `${accepted} accepted`,
      state: check(accepted >= 2),
      fix: { kind: "tab", tab: "registration" },
    },
  ];

  if (!matches) {
    items.push(bracketUnavailable("bracket", "Bracket defined and stage 1 seeded", matches));
    return items;
  }

  const hasChampionship = matches.some((match) => !match.targetMatchId);
  const stage1 = matches.filter((match) => match.stageLevel === 1);
  const stage1Capacity = stage1.reduce((sum, match) => sum + match.teamCount, 0);
  const seeded = stage1.reduce((sum, match) => sum + (match.entries?.length ?? 0), 0);
  const defined = matches.length > 0 && hasChampionship;

  items.push(
    {
      id: "bracket-defined",
      label: "Bracket fits the accepted teams",
      detail: !defined
        ? "No complete bracket yet"
        : `Stage 1 holds ${plural(stage1Capacity, "team")}, ${accepted} accepted`,
      state: check(defined && stage1.length > 0 && stage1Capacity === accepted),
      fix: { kind: "tab", tab: "bracket" },
    },
    {
      id: "bracket-seeded",
      label: "Stage 1 seeded",
      detail: `${seeded} of ${accepted} accepted teams placed`,
      state: check(defined && seeded === accepted),
      fix: { kind: "tab", tab: "bracket" },
    }
  );
  return items;
}

export function completionRequirements(matches: Remote<Match>): Requirement[] {
  if (!matches) {
    return [
      bracketUnavailable("matches-complete", "All matches completed", matches),
      bracketUnavailable("champion", "Champion decided", matches),
    ];
  }
  const completed = matches.filter((match) => match.isCompleted).length;
  const final = matches.find((match) => !match.targetMatchId);
  const champions =
    final?.entries?.filter((entry) => entry.finalRank === 1) ?? [];

  return [
    {
      id: "matches-complete",
      label: "All matches completed",
      detail: matches.length
        ? `${completed} of ${plural(matches.length, "match")} completed`
        : "No bracket defined",
      state: check(matches.length > 0 && completed === matches.length),
      fix: { kind: "tab", tab: "bracket" },
    },
    {
      id: "champion",
      label: "Champion decided",
      detail:
        champions.length === 1
          ? `${champions[0].teamName ?? "Winner"} ranked 1st in the final`
          : champions.length > 1
            ? `${champions.length} teams ranked 1st in the final`
            : "The final has no 1st place yet",
      state: check(champions.length === 1),
      fix: { kind: "tab", tab: "bracket" },
    },
  ];
}
