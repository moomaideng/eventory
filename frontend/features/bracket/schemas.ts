import { z } from "zod";

export const bracketMatchNodeSchema = z.object({
  tempId: z.string().min(1, "Identifier is required"),
  name: z.string().trim().min(1, "Match name is required").max(160, "Name too long"),
  venue: z.string().trim().min(1, "Venue is required").max(160, "Venue too long"),
  stageLevel: z.coerce.number().int().min(1, "Stage level must be at least 1"),
  teamCount: z.coerce.number().int().min(1, "Team count must be at least 1"),
  advancingCount: z.coerce.number().int().min(0, "Advancing count cannot be negative"),
  targetTempId: z.string().optional().nullable(),
});

export type BracketMatchNode = z.infer<typeof bracketMatchNodeSchema>;

export const teamPlacementSchema = z.object({
  teamId: z.string().uuid("Invalid team ID"),
  finalRank: z.coerce.number().int().min(1, "Rank must be at least 1"),
  score: z.string().trim().max(32, "Score must be under 32 characters").optional(),
});

export type TeamPlacementInput = z.infer<typeof teamPlacementSchema>;

export const recordMatchResultsSchema = z.object({
  placements: z.array(teamPlacementSchema).min(1, "At least one placement is required"),
});

export type RecordMatchResultsInput = z.infer<typeof recordMatchResultsSchema>;

export const stage1SeedAssignmentSchema = z.object({
  matchId: z.string().uuid("Invalid match ID"),
  teamIds: z.array(z.string().uuid("Invalid team ID")),
});

export type Stage1SeedAssignment = z.infer<typeof stage1SeedAssignmentSchema>;
