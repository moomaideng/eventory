import { z } from "zod";

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
