import { z } from "zod";

/**
 * Validates the sponsor profile form (#63) before sending to
 * upsert-my-sponsor-profile. Mirrors UpsertSponsorProfileRequest.
 */
export const sponsorProfileSchema = z.object({
  sponsorName: z
    .string()
    .trim()
    .min(1, "Enter a sponsor name.")
    .max(120, "Sponsor name must be 120 characters or fewer."),
  sponsorEmail: z
    .string()
    .trim()
    .transform((val) => (val === "" ? undefined : val))
    .pipe(z.string().email("Enter a valid email address.").optional()),
});

export type SponsorProfileInput = z.infer<typeof sponsorProfileSchema>;
