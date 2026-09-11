import { z } from "zod";
import type { components } from "@/lib/api/schema";

export type OrganizerSummary =
  components["schemas"]["OrganizerTournamentSummary"];
export type DashboardEntry = components["schemas"]["OrganizerDashboardEntry"];

/** The lifecycle statuses the API accepts, taken from the generated schema. */
export type TournamentStatus =
  components["schemas"]["OverrideTournamentStatusInputBody"]["status"];

export const dashboardQueryOptions = {
  retry: false,
  staleTime: 0,
  refetchOnWindowFocus: true,
  refetchInterval: 30_000,
} as const;

export function statusLabel(status: string) {
  return status.toLowerCase().replaceAll("_", " ");
}

// Mirrors the server's limits in internal/usecases/tournament_status_usecase.go
// so the organizer sees the problem before a request is sent. The server
// re-validates regardless; this is convenience, not enforcement.
export const REASON_MIN_LENGTH = 10;
export const REASON_MAX_LENGTH = 500;

export const statusOverrideSchema = z.object({
  status: z
    .string({ message: "Choose a status." })
    .min(1, "Choose a status."),
  reason: z
    .string()
    .trim()
    .min(REASON_MIN_LENGTH, `Give at least ${REASON_MIN_LENGTH} characters.`)
    .max(REASON_MAX_LENGTH, `Keep it under ${REASON_MAX_LENGTH} characters.`),
});

export type StatusOverrideInput = z.infer<typeof statusOverrideSchema>;

// Statuses a tournament cannot be moved out of again. Shown as a confirmation
// before submitting, because the API will not let the organizer take it back.
//
// TODO(US6-x): warn when completing a tournament that has no recorded winner.
// There is no winner, placement, or bracket model yet, so the copy below is
// deliberately unconditional.
export const TERMINAL_WARNING: Record<string, string> = {
  COMPLETED:
    "A completed tournament becomes a read-only archive. You will not be able to reopen it.",
  CANCELLED:
    "Cancelling refunds every entrant and sponsor. You will not be able to revive this tournament.",
};
