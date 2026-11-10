import { extractProblemMessage } from "@/lib/schemas";
import type { components } from "@/lib/api/schema";

export type Tournament = components["schemas"]["TournamentResponse"];

export const PAGE_SIZE = 6;

const tournamentDateFormatter = new Intl.DateTimeFormat("en-GB", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "Asia/Bangkok",
});

const currencyFormatters = new Map<string, Intl.NumberFormat>();

export function formatTournamentDate(value?: string | null) {
  if (!value) return "Date not set";
  return tournamentDateFormatter.format(new Date(value));
}

export function formatDateRange(startAt?: string | null, endAt?: string | null) {
  if (!startAt && !endAt) return "Dates not set";
  if (!startAt) return `Until ${formatTournamentDate(endAt)}`;
  if (!endAt) return `From ${formatTournamentDate(startAt)}`;
  return `${formatTournamentDate(startAt)} – ${formatTournamentDate(endAt)}`;
}

export function formatMoney(amount: number, currency: string, freeLabel = false) {
  if (freeLabel && amount === 0) return "Free entry";
  const safeCurrency = currency?.trim() || "THB";
  let formatter = currencyFormatters.get(safeCurrency);
  if (!formatter) {
    try {
      formatter = new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: safeCurrency,
        maximumFractionDigits: 0,
      });
      currencyFormatters.set(safeCurrency, formatter);
    } catch {
      return `${amount.toLocaleString("en-US")} ${currency}`;
    }
  }
  return formatter.format(amount);
}

export function formatEntryFee(entryFee: number, currency: string) {
  return formatMoney(entryFee, currency, true);
}

const percentageFormatter = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 1,
});

export function formatPercentage(value: number) {
  return `${percentageFormatter.format(value)}%`;
}

export function formatRegistrationType(value: string) {
  if (value === "TEAM") return "Team registration";
  return "Solo registration";
}

export function formatTournamentStatus(
  value: string,
  registrationDeadline?: string | null
) {
  switch (value) {
    case "DRAFT":
      return "Draft";
    case "CROWDFUNDING":
      return "Crowdfunding";
    case "REGISTRATION":
      if (
        registrationDeadline &&
        new Date(registrationDeadline).getTime() <= Date.now()
      ) {
        return "Registration closed";
      }
      return "Registration open";
    case "ONGOING":
      return "Ongoing";
    case "COMPLETED":
      return "Completed";
    default:
      return value.replaceAll("_", " ").toLowerCase();
  }
}

export function formatRegistrationStatus(
  status: string,
  deadline?: string | null
) {
  switch (status) {
    case "REGISTRATION":
      if (deadline) {
        const isClosed = new Date(deadline).getTime() <= Date.now();
        if (isClosed) return "Registration closed";
        return `Registration closes ${formatTournamentDate(deadline)}`;
      }
      return "Registration open";
    case "ONGOING":
      return "Tournament in progress";
    case "COMPLETED":
      return "Tournament completed";
    case "CROWDFUNDING":
      return "Funding in progress";
    case "DRAFT":
      return "Draft";
    default:
      return deadline
        ? `Registration closes ${formatTournamentDate(deadline)}`
        : formatTournamentStatus(status, deadline);
  }
}

/**
 * Registration is a derived substate of REGISTRATION: open until the
 * registration deadline passes. A missing deadline never counts as open.
 */
export function isRegistrationOpen(
  tournament: { status: string; registrationDeadline?: string | null },
  now: number = Date.now()
) {
  if (tournament.status !== "REGISTRATION" || !tournament.registrationDeadline) {
    return false;
  }
  return new Date(tournament.registrationDeadline).getTime() > now;
}

export function initials(name: string) {
  return name
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
}

const dateFormatter = new Intl.DateTimeFormat("en-GB", {
  dateStyle: "medium",
  timeZone: "Asia/Bangkok",
});

export function formatDate(value: string) {
  return dateFormatter.format(new Date(value));
}

export function problemMessage(error: unknown, fallback: string) {
  return extractProblemMessage(error, fallback);
}
