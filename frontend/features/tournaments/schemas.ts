import { z } from "zod";

/**
 * Validates a tournament UUID path parameter.
 */
export const tournamentIdParamSchema = z.uuid("Invalid tournament ID format.");

export type TournamentIdParam = z.infer<typeof tournamentIdParamSchema>;

/**
 * Validates form submission for creating a new team lobby in a tournament.
 */
export const createTeamLobbySchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Choose a team name before creating the lobby.")
    .max(120, "Team name must be 120 characters or fewer."),
});

export type CreateTeamLobbyInput = z.infer<typeof createTeamLobbySchema>;

/**
 * Validates and sanitizes tournament catalog query filters.
 */
export const tournamentFiltersSchema = z.object({
  q: z.string().trim().max(100).optional().catch(undefined),
  startFrom: z.string().trim().optional().catch(undefined),
  startTo: z.string().trim().optional().catch(undefined),
  maxEntryFee: z.coerce.number().min(0).optional().catch(undefined),
  page: z.coerce.number().int().positive().default(1).catch(1),
});

export type TournamentFilters = z.infer<typeof tournamentFiltersSchema>;
