import { z } from "zod";
import { optionalPhoneSchema } from "@/lib/schemas";
import { normalizeSponsorTags } from "@/features/sponsor/tags";

export const sponsorTagsSchema = z
  .array(z.string())
  .transform(normalizeSponsorTags)
  .pipe(
    z
      .array(z.string().max(40, "Each tag must be 40 characters or fewer."))
      .min(1, "Enter at least one tag.")
      .max(10, "Use up to 10 tags.")
  );

const optionalHttpUrlSchema = (label: string) =>
  z
    .string()
    .trim()
    .max(2048, `${label} must be 2048 characters or fewer.`)
    .refine((value) => {
      if (value === "") return true;
      try {
        const parsed = new URL(value);
        return (
          (parsed.protocol === "http:" || parsed.protocol === "https:") &&
          Boolean(parsed.hostname)
        );
      } catch {
        return false;
      }
    }, `Enter a valid http:// or https:// ${label.toLowerCase()}.`)
    .transform((value) => (value === "" ? undefined : value));

export const sponsorProfileSchema = z.object({
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
  phone: optionalPhoneSchema.optional(),
  websiteUrl: optionalHttpUrlSchema("Brand website URL").optional(),
  logoUrl: optionalHttpUrlSchema("Logo image URL").optional(),
  tags: sponsorTagsSchema,
  privacyConsentAccepted: z.boolean().optional(),
});

export const sponsorCreateSchema = sponsorProfileSchema.extend({
  privacyConsentAccepted: z.literal(true, {
    message:
      "Please confirm consent to store and display your sponsor profile.",
  }),
});

export const sponsorOnboardingSchema = sponsorCreateSchema;
export type SponsorOnboardingValues = z.infer<typeof sponsorOnboardingSchema>;
export type SponsorProfileInput = z.infer<typeof sponsorProfileSchema>;
