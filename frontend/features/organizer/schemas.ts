import { z } from "zod";

/**
 * Validates the organizer profile form (#62) before sending to
 * upsert-my-organizer-profile. Mirrors UpsertOrganizerProfileRequest.
 */
export const organizerProfileSchema = z.object({
  organizerName: z
    .string()
    .trim()
    .min(1, "Enter an organizer name.")
    .max(150, "Organizer name must be 150 characters or fewer."),
  organizerEmail: z
    .string()
    .trim()
    .transform((val) => (val === "" ? undefined : val))
    .pipe(z.string().email("Enter a valid email address.").optional()),
});

export type OrganizerProfileInput = z.infer<typeof organizerProfileSchema>;
