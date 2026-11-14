import { z } from "zod";

/**
 * Validates an organizer account UUID path parameter.
 */
export const organizerIdParamSchema = z.uuid("Invalid organizer ID format.");

export type OrganizerIdParam = z.infer<typeof organizerIdParamSchema>;

export const ORGANIZER_TOURNAMENT_STATUSES = [
  "ALL",
  "CROWDFUNDING",
  "REGISTRATION",
  "ONGOING",
  "COMPLETED",
] as const;

export type OrganizerTournamentStatusFilter =
  (typeof ORGANIZER_TOURNAMENT_STATUSES)[number];

/**
 * Validates URL search params for an organizer's hosted tournament list.
 */
export const organizerTournamentsQuerySchema = z.object({
  status: z.enum(ORGANIZER_TOURNAMENT_STATUSES).default("ALL").catch("ALL"),
  page: z.coerce.number().int().positive().default(1).catch(1),
});

export type OrganizerTournamentsQuery = z.infer<
  typeof organizerTournamentsQuerySchema
>;
