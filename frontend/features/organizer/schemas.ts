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
    registrationDeadline: z
      .string()
      .min(1, "Registration deadline is required."),
    startAt: z.string().min(1, "Tournament start time is required."),
    endAt: z.string().min(1, "Tournament end time is required."),
  })
  .superRefine((data, ctx) => {
    if (data.registrationMode === "TEAM" && data.minTeamSize > data.maxTeamSize) {
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

/**
 * Validates registration form builder schemas (#63).
 */
export const fieldTypeSchema = z.enum([
  "SHORT_TEXT",
  "LONG_TEXT",
  "NUMBER",
  "SELECT_ONE",
  "SELECT_MULTIPLE",
]);

export type FieldType = z.infer<typeof fieldTypeSchema>;

export const formFieldOptionSchema = z.object({
  id: z.string().min(1),
  label: z.string().trim().min(1, "Option label is required."),
});

export type FormFieldOption = z.infer<typeof formFieldOptionSchema>;

export const formFieldConfigSchema = z.object({
  minLength: z.coerce.number().int().min(0).optional(),
  maxLength: z.coerce.number().int().min(0).optional(),
  min: z.coerce.number().optional(),
  max: z.coerce.number().optional(),
  allowDecimals: z.boolean().optional(),
  options: z.array(formFieldOptionSchema).optional(),
  minSelections: z.coerce.number().int().min(0).optional(),
  maxSelections: z.coerce.number().int().min(0).optional(),
});

export type FormFieldConfig = z.infer<typeof formFieldConfigSchema>;

export const formFieldSchema = z.object({
  id: z.string().min(1),
  type: fieldTypeSchema,
  label: z.string().trim().min(1, "Field label is required."),
  description: z.string().trim().optional(),
  required: z.boolean().default(false),
  order: z.number().int().default(1),
  config: formFieldConfigSchema.default({}),
});

export type FormField = z.infer<typeof formFieldSchema>;

export const registrationFormSchema = z.object({
  title: z
    .string()
    .trim()
    .min(1, "Form title is required.")
    .max(160, "Form title must be 160 characters or fewer."),
  description: z
    .string()
    .trim()
    .max(2000, "Description must be 2,000 characters or fewer.")
    .optional(),
  fields: z.array(formFieldSchema),
});

export type RegistrationFormValues = z.infer<typeof registrationFormSchema>;

