import { z } from "zod";

export const sponsorTagsSchema = z
  .array(z.string())
  .transform((tags) => [
    ...new Set(tags.map((tag) => tag.trim().toLowerCase()).filter(Boolean)),
  ])
  .pipe(
    z
      .array(z.string().max(40, "Each tag must be 40 characters or fewer."))
      .min(1, "Enter at least one tag.")
      .max(10, "Use up to 10 tags.")
  );

export const sponsorOnboardingSchema = z.object({
  sponsorName: z
    .string()
    .trim()
    .min(1, "Enter a sponsor name.")
    .max(120, "Sponsor name must be 120 characters or fewer."),
  sponsorEmail: z
    .string()
    .trim()
    .email("Enter a valid contact email.")
    .max(255),
  tags: sponsorTagsSchema,
});
export const sponsorProfileSchema = sponsorOnboardingSchema;
export type SponsorOnboardingValues = z.infer<typeof sponsorOnboardingSchema>;
export type SponsorProfileInput = z.infer<typeof sponsorProfileSchema>;
