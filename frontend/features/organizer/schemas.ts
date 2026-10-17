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
 * Validates form submission when hosting a new tournament.
 * Enforces date sequencing (registrationDeadline < startAt < endAt)
 * and roster sizing invariants.
 */
export const createTournamentFormSchema = z
  .object({
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
      .max(4000, "Description must be 4,000 characters or fewer.")
      .default(""),
    registrationMode: z.enum(["SOLO", "TEAM"]),
    minTeamSize: z.coerce
      .number()
      .int("Team size must be a whole number.")
      .min(1, "Minimum team size must be at least 1.")
      .default(1),
    maxTeamSize: z.coerce
      .number()
      .int("Team size must be a whole number.")
      .min(1, "Maximum team size must be at least 1.")
      .default(1),
    capacity: z.coerce
      .number()
      .int("Capacity must be a whole number.")
      .min(1, "Capacity must be at least 1.")
      .default(16),
    entryFee: z.coerce
      .number()
      .int("Entry fee must be a whole number.")
      .min(0, "Entry fee cannot be negative.")
      .default(0),
    crowdfunding: z.boolean(),
    fundingGoal: z.coerce.number().int("Funding goal must be a whole number."),
    fundingDeadline: z.string(),
    registrationDeadline: z
      .string()
      .min(1, "Registration deadline is required."),
    startAt: z.string().min(1, "Tournament start time is required."),
    endAt: z.string().min(1, "Tournament end time is required."),
  })
  .superRefine((data, ctx) => {
    if (data.crowdfunding) {
      const fundingDeadline = new Date(data.fundingDeadline).getTime();
      const registrationDeadline = new Date(
        data.registrationDeadline
      ).getTime();
      if (data.fundingGoal <= 0) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: "Enter a funding goal above 0.",
          path: ["fundingGoal"],
        });
      }
      if (
        !data.fundingDeadline ||
        Number.isNaN(fundingDeadline) ||
        fundingDeadline <= Date.now() ||
        (!Number.isNaN(registrationDeadline) &&
          fundingDeadline >= registrationDeadline)
      ) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message:
            "Choose a future funding deadline before registration closes.",
          path: ["fundingDeadline"],
        });
      }
    }
    if (
      data.registrationMode === "TEAM" &&
      data.minTeamSize > data.maxTeamSize
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Maximum team size must be at least minimum team size.",
        path: ["maxTeamSize"],
      });
    }

    const deadline = new Date(data.registrationDeadline).getTime();
    const start = new Date(data.startAt).getTime();
    const end = new Date(data.endAt).getTime();

    if (Number.isNaN(deadline)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Please select a valid registration deadline.",
        path: ["registrationDeadline"],
      });
    }

    if (Number.isNaN(start)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Please select a valid start date and time.",
        path: ["startAt"],
      });
    }

    if (Number.isNaN(end)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Please select a valid end date and time.",
        path: ["endAt"],
      });
    }

    if (!Number.isNaN(deadline) && !Number.isNaN(start) && deadline >= start) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Registration deadline must be before tournament start.",
        path: ["registrationDeadline"],
      });
    }

    if (!Number.isNaN(start) && !Number.isNaN(end) && start >= end) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Tournament end time must be after tournament start.",
        path: ["endAt"],
      });
    }
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
      description: z.string().trim().max(4000),
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
