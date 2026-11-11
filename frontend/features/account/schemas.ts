import { z } from "zod";
import { onboardingFormSchema } from "@/features/auth/schemas";

/**
 * Validates the primary account settings form (US1-4).
 * Same display name and handle rules as onboarding. Phone numbers live on the
 * organizer and sponsor profiles, and the avatar is the sign-in photo.
 */
export const accountSettingsSchema = onboardingFormSchema;

export type AccountSettingsInput = z.infer<typeof accountSettingsSchema>;
