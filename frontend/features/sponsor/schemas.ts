import { z } from "zod";
import { sponsorCategories } from "./categories";

/**
 * Validates the sponsor profile form (#63) before sending to
 * upsert-my-sponsor-profile. Mirrors UpsertSponsorProfileRequest.
 */
export const sponsorOnboardingSchema = z.object({
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

export type SponsorOnboardingValues = z.infer<typeof sponsorOnboardingSchema>;

export const sponsorProfileSchema = sponsorOnboardingSchema
  .extend({
    category: z.union([z.enum(sponsorCategories), z.literal("")]).optional(),
    directoryVisible: z.boolean().optional(),
  })
  .refine(
    (value) =>
      !value.directoryVisible ||
      Boolean(value.category && value.sponsorEmail),
    {
      message: "Choose a category and enter a contact email before joining the directory.",
      path: ["directoryVisible"],
    }
  );

export type SponsorProfileInput = z.infer<typeof sponsorProfileSchema>;
