import { z } from "zod";

/**
 * Validates OAuth callback query parameters from the auth provider redirect.
 */
export const authCallbackQuerySchema = z.object({
  code: z.string().trim().min(1).optional(),
  error: z.string().optional(),
  error_description: z.string().optional(),
});

export type AuthCallbackQuery = z.infer<typeof authCallbackQuerySchema>;

/**
 * Validates Supabase user metadata structure on successful OAuth exchange.
 */
export const authUserMetadataSchema = z
  .object({
    full_name: z.string().trim().min(1).optional(),
    name: z.string().trim().min(1).optional(),
    avatar_url: z.string().url().optional(),
    picture: z.string().url().optional(),
  })
  .passthrough();

export type AuthUserMetadata = z.infer<typeof authUserMetadataSchema>;

/**
 * Validates JIT account auto-provisioning payload before posting to backend API.
 */
export const accountProvisionSchema = z.object({
  displayName: z.string().trim().min(1).max(100),
  avatarUrl: z.string().url().optional(),
});

export type AccountProvisionInput = z.infer<typeof accountProvisionSchema>;
