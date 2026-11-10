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

export const editGeneralSectionSchema = createTournamentFormSchema;
export type EditGeneralSectionInput = z.infer<typeof editGeneralSectionSchema>;

export const editGateSectionSchema = z
  .object({
    registrationMode: z.enum(["SOLO", "TEAM"]),
    minTeamSize: z.coerce
      .number()
      .int("Team size must be a whole number.")
      .min(1, "Minimum team size must be at least 1."),
    maxTeamSize: z.coerce
      .number()
      .int("Team size must be a whole number.")
      .min(1, "Maximum team size must be at least 1."),
    capacity: z.coerce
      .number()
      .int("Capacity must be a whole number.")
      .min(1, "Capacity must be at least 1."),
    entryFee: z.coerce
      .number()
      .int("Entry fee must be a whole number.")
      .min(0, "Entry fee cannot be negative."),
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
export type EditGateSectionInput = z.infer<typeof editGateSectionSchema>;

export const editTimelineSectionSchema = z
  .object({
    registrationDeadline: z.string().default(""),
    startAt: z.string().default(""),
    endAt: z.string().default(""),
  })
  .superRefine((data, ctx) => {
    const deadline = data.registrationDeadline
      ? new Date(data.registrationDeadline).getTime()
      : null;
    const start = data.startAt ? new Date(data.startAt).getTime() : null;
    const end = data.endAt ? new Date(data.endAt).getTime() : null;

    if (deadline && start && deadline >= start) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Registration deadline must be before tournament start.",
        path: ["registrationDeadline"],
      });
    }

    if (start && end && start >= end) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Tournament end time must be after tournament start.",
        path: ["endAt"],
      });
    }
  });
export type EditTimelineSectionInput = z.infer<
  typeof editTimelineSectionSchema
>;

export const editCrowdfundingSectionSchema = z
  .object({
    fundingGoal: z.coerce
      .number()
      .int("Funding goal must be a whole number.")
      .min(1, "Funding goal must be greater than 0."),
    fundingDeadline: z.string().min(1, "Funding deadline is required."),
  })
  .superRefine((data, ctx) => {
    const deadline = new Date(data.fundingDeadline).getTime();
    if (Number.isNaN(deadline) || deadline <= Date.now()) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Funding deadline must be in the future.",
        path: ["fundingDeadline"],
      });
    }
  });
export type EditCrowdfundingSectionInput = z.infer<
  typeof editCrowdfundingSectionSchema
>;
