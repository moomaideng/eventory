import type { components } from "@/lib/api/schema";

export type OrganizerSummary =
  components["schemas"]["OrganizerTournamentSummary"];
export type DashboardEntry = components["schemas"]["OrganizerDashboardEntry"];


export const dashboardQueryOptions = {
  retry: false,
  staleTime: 0,
  refetchOnWindowFocus: true,
  refetchInterval: 30_000,
} as const;

export function toLocalInput(value?: string | null) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Date(date.getTime() - date.getTimezoneOffset() * 60_000)
    .toISOString()
    .slice(0, 16);
}

export function statusLabel(status: string) {
  if (status === "LOCKED") return "Waiting for approval";
  if (status === "PAYMENT_PENDING") return "Awaiting payment";
  return status.toLowerCase().replaceAll("_", " ");
}

export function statusTitle(status: string) {
  return statusLabel(status)
    .split(" ")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

export const REASON_MAX_LENGTH = 500;

const relativeTimeFormatter = new Intl.RelativeTimeFormat("en", {
  numeric: "auto",
});

const MINUTE_MS = 60_000;
const HOUR_MS = 60 * MINUTE_MS;
const DAY_MS = 24 * HOUR_MS;
const MONTH_MS = 30 * DAY_MS;

/** Human relative time such as "in 11 days" or "3 hours ago"; null when missing. */
export function formatRelative(value?: string | null, now: number = Date.now()) {
  if (!value) return null;
  const time = new Date(value).getTime();
  if (Number.isNaN(time)) return null;
  const diff = time - now;
  const abs = Math.abs(diff);
  if (abs < HOUR_MS) {
    return relativeTimeFormatter.format(Math.round(diff / MINUTE_MS), "minute");
  }
  if (abs < DAY_MS) {
    return relativeTimeFormatter.format(Math.round(diff / HOUR_MS), "hour");
  }
  if (abs < MONTH_MS) {
    return relativeTimeFormatter.format(Math.round(diff / DAY_MS), "day");
  }
  return relativeTimeFormatter.format(Math.round(diff / MONTH_MS), "month");
}
