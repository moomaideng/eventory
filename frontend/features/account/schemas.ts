import { z } from "zod";
import { onboardingFormSchema } from "@/features/auth/schemas";

// Only http(s) URLs: blocks javascript:, data:, file: and similar schemes.
const httpUrl = z.url({ protocol: /^https?$/ });

// Optional. The API does not validate phone numbers, so this only rejects
// obvious junk (letters, symbols) and keeps real formats such as
// "+66 81 234 5678" and "081-234-5678".
const PHONE_PATTERN = /^\+?[0-9][0-9\s()-]{5,}$/;

/**
 * Validates the primary account settings form (US1-4).
 * Reuses the handle and display name rules from onboarding. Empty phone or
 * avatar is allowed: an empty string tells PATCH /accounts/me to clear it.
 */
export const accountSettingsSchema = onboardingFormSchema.extend({
  phone: z
    .string()
    .trim()
    .max(20, "Phone number must be 20 characters or fewer.")
    .refine(
      (value) => value === "" || PHONE_PATTERN.test(value),
      "Enter a valid phone number, e.g. +66 81 234 5678."
    ),
  avatarUrl: z
    .string()
    .trim()
    .max(2048, "Avatar URL must be 2048 characters or fewer.")
    .refine(
      (value) => value === "" || httpUrl.safeParse(value).success,
      "Enter a valid http(s) image URL."
    ),
});

export type AccountSettingsInput = z.infer<typeof accountSettingsSchema>;
