import { z } from "zod";

/**
 * Schema for entering a 6-character alphanumeric lobby invite code in the join form.
 */
export const lobbyCodeSchema = z
  .string()
  .trim()
  .toUpperCase()
  .regex(/^[A-Z0-9]{6}$/, "Invalid lobby invite code.");

export type LobbyCode = z.infer<typeof lobbyCodeSchema>;

export const lobbyInviteParamSchema = z
  .string()
  .trim()
  .toUpperCase()
  .regex(/^[A-Z0-9]{6}$/, "Invalid lobby invite code.");

export type LobbyInviteParam = z.infer<typeof lobbyInviteParamSchema>;
