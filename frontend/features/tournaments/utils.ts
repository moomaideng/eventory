import type { components } from "@/lib/api/schema";

export type Tournament = components["schemas"]["TournamentResponse"];

export const PAGE_SIZE = 6;

const tournamentDateFormatter = new Intl.DateTimeFormat("en-GB", {
  dateStyle: "medium",
  timeStyle: "short",
  timeZone: "Asia/Bangkok",
});

const currencyFormatters = new Map<string, Intl.NumberFormat>();

export function formatTournamentDate(value: string) {
  return tournamentDateFormatter.format(new Date(value));
}

export function formatDateRange(startAt: string, endAt: string) {
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
  if (value === "BOTH") return "Solo or team registration";
  return "Solo registration";
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
  if (
    error &&
    typeof error === "object" &&
    "detail" in error &&
    typeof error.detail === "string"
  ) {
    return error.detail;
  }
  return fallback;
}


