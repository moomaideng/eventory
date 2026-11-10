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

/**
 * Validates onboarding form inputs for organizer profile setup.
 */
export const organizerOnboardingSchema = organizerProfileSchema;
export type OrganizerOnboardingValues = OrganizerProfileInput;

/**
 * Validates form submission when creating a new draft tournament.
 * In the refactored tournament lifecycle, all tournaments initialize in DRAFT.
 * Only basic tournament info is required upfront; registration gate,
 * timeline, and funding settings are configured in draft state.
 */
export const createTournamentFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Tournament name is required.")
    .max(160, "Tournament name must be 160 characters or fewer."),
  game: z
    .string()
    .trim()
    .min(1, "Game title is required.")
    .max(80, "Game title must be 80 characters or fewer."),
  location: z
    .string()
    .trim()
    .min(1, "Location is required (e.g. 'Online' or venue address).")
    .max(160, "Location must be 160 characters or fewer."),
  description: z
    .string()
    .trim()
    .min(1, "Short description is required.")
    .max(500, "Short description must be 500 characters or fewer."),
  detail: z
    .string()
    .trim()
    .max(10000, "Tournament details & rules must be 10,000 characters or fewer.")
    .default(""),
});

export type CreateTournamentFormInput = z.infer<
  typeof createTournamentFormSchema
>;

export function fundingSettingsSchema(registrationDeadline: string) {
  return z
    .object({
      fundingGoal: z.coerce
        .number()
        .int("Funding goal must be a whole number.")
        .positive("Enter a funding goal above 0."),
      fundingDeadline: z.string().min(1, "Choose a funding deadline."),
    })
    .superRefine((data, ctx) => {
      const deadline = new Date(data.fundingDeadline).getTime();
      if (
        Number.isNaN(deadline) ||
        deadline <= Date.now() ||
        deadline >= new Date(registrationDeadline).getTime()
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message:
            "Choose a future funding deadline before registration closes.",
          path: ["fundingDeadline"],
        });
      }
    });
}

export function plannedEventDetailsSchema(fundingDeadline: string) {
  return z
    .object({
      location: z.string().trim().min(1, "Enter a planned location.").max(160),
      description: z.string().trim().max(500, "Short description must be 500 characters or fewer.").default(""),
      detail: z.string().trim().max(10000, "Tournament details & rules must be 10,000 characters or fewer.").default(""),
      registrationDeadline: z
        .string()
        .min(1, "Choose a registration deadline."),
      startAt: z.string().min(1, "Choose a planned start time."),
      endAt: z.string().min(1, "Choose a planned end time."),
    })
    .superRefine((data, ctx) => {
      const funding = new Date(fundingDeadline).getTime();
      const registration = new Date(data.registrationDeadline).getTime();
      const start = new Date(data.startAt).getTime();
      const end = new Date(data.endAt).getTime();
      if (Number.isNaN(registration) || registration <= funding) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Registration must close after the funding deadline.",
          path: ["registrationDeadline"],
        });
      }
      if (Number.isNaN(start) || start <= registration) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "The event must start after registration closes.",
          path: ["startAt"],
        });
      }
      if (Number.isNaN(end) || end <= start) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "The event must end after it starts.",
          path: ["endAt"],
        });
      }
    });
}

export const registrationRulesSchema = z
  .object({
    registrationMode: z.enum(["SOLO", "TEAM"]),
    minTeamSize: z.coerce.number().int().min(1),
    maxTeamSize: z.coerce.number().int().min(1),
    capacity: z.coerce.number().int().min(1, "Capacity must be at least 1."),
    entryFee: z.coerce.number().int().min(0, "Entry fee cannot be negative."),
  })
  .superRefine((data, ctx) => {
    if (
      data.registrationMode === "TEAM" &&
      data.maxTeamSize < data.minTeamSize
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Maximum team size must be at least minimum team size.",
        path: ["maxTeamSize"],
      });
    }
  });
