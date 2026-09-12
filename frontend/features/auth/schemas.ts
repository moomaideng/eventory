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
export const authUserMetadataSchema = z.looseObject({
  full_name: z.string().trim().min(1).optional(),
  name: z.string().trim().min(1).optional(),
  avatar_url: z.url().optional(),
  picture: z.url().optional(),
});

export type AuthUserMetadata = z.infer<typeof authUserMetadataSchema>;

/**
 * Validates JIT account auto-provisioning payload before posting to backend API.
 */
export const accountProvisionSchema = z.object({
  displayName: z.string().trim().min(1).max(100),
  avatarUrl: z.url().optional(),
});

export type AccountProvisionInput = z.infer<typeof accountProvisionSchema>;

/**
 * Validates onboarding form inputs for user display name and unique handle.
 */
export const onboardingFormSchema = z.object({
  displayName: z
    .string()
    .trim()
    .min(1, "Display name cannot be empty")
    .max(64, "Display name must be 64 characters or fewer"),
  handle: z
    .string()
    .trim()
    .toLowerCase()
    .min(3, "Handle must be at least 3 characters")
    .max(32, "Handle must be 32 characters or fewer")
    .regex(
      /^[a-z0-9_]+$/,
      "Handle can only contain lowercase letters, numbers, and underscores"
    ),
});

export type OnboardingFormValues = z.infer<typeof onboardingFormSchema>;
