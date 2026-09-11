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

export function statusLabel(status: string) {
  return status.toLowerCase().replaceAll("_", " ");
}
