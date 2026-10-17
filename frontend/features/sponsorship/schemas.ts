import { z } from "zod";

/**
 * Validates the sponsorship tier form (US5-4). Mirrors SponsorshipTierBody;
 * perks arrive as one textarea and are split into one perk per line.
 */
export const sponsorshipTierSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Enter a tier name.")
    .max(80, "Tier name must be 80 characters or fewer."),
  price: z.coerce
    .number({ message: "Enter a price." })
    .int("Price must be a whole number.")
    .min(1, "Price must be greater than zero."),
  slots: z.coerce
    .number({ message: "Enter the number of slots." })
    .int("Slots must be a whole number.")
    .min(1, "A tier needs at least one slot."),
  perks: z
    .string()
    .transform((value) =>
      value
        .split("\n")
        .map((perk) => perk.trim())
        .filter(Boolean)
    )
    .pipe(
      z
        .array(
          z.string().max(200, "Each perk must be 200 characters or fewer.")
        )
        .min(1, "Add at least one perk.")
        .max(10, "A tier can have at most 10 perks.")
    ),
  sortOrder: z.coerce.number().int().default(0),
});

export type SponsorshipTierFormInput = z.input<typeof sponsorshipTierSchema>;
export type SponsorshipTierPayload = z.output<typeof sponsorshipTierSchema>;

export const sponsorshipBrandSchema = z.object({
  websiteUrl: z
    .string()
    .trim()
    .max(2048)
    .url("Enter a complete website URL.")
    .refine((value) => {
      try {
        const url = new URL(value);
        return (
          ["https:", "http:"].includes(url.protocol) &&
          !url.username &&
          !url.password
        );
      } catch {
        return false;
      }
    }, "Use an HTTP or HTTPS website without login details."),
  logo: z
    .file()
    .min(1, "Choose a logo.")
    .max(1024 * 1024, "Logo must be 1 MB or smaller.")
    .mime(["image/png", "image/jpeg"], "Choose a PNG or JPEG logo."),
});

export const sponsorshipCheckoutSchema = z.object({
  id: z.uuid(),
  tournamentId: z.uuid(),
  status: z.enum(["PENDING", "CONFIRMED", "CANCELLED", "EXPIRED"]),
  paymentId: z.uuid().nullish(),
});
