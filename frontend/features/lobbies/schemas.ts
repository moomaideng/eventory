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

/**
 * Schema for validating the inviteCode route parameter from URL path.
 * Supports standard 6-character codes as well as seeded format codes (up to 32 chars).
 */
export const lobbyInviteParamSchema = z
  .string()
  .trim()
  .toUpperCase()
  .regex(/^[A-Z0-9_-]{3,32}$/, "Invalid lobby invite code.");

export type LobbyInviteParam = z.infer<typeof lobbyInviteParamSchema>;
