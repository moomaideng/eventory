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
