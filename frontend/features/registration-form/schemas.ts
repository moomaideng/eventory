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

export const textConfigSchema = z.object({
  minLength: z.coerce.number().int().min(0).optional(),
  maxLength: z.coerce.number().int().min(0).optional(),
});
export type TextConfig = z.infer<typeof textConfigSchema>;

export const numberConfigSchema = z.object({
  min: z.coerce.number().optional(),
  max: z.coerce.number().optional(),
  allowDecimals: z.boolean().optional(),
});
export type NumberConfig = z.infer<typeof numberConfigSchema>;

export const selectOneConfigSchema = z.object({
  options: z.array(formFieldOptionSchema).default([]),
});
export type SelectOneConfig = z.infer<typeof selectOneConfigSchema>;

export const selectMultipleConfigSchema = z.object({
  options: z.array(formFieldOptionSchema).default([]),
  minSelections: z.coerce.number().int().min(0).optional(),
  maxSelections: z.coerce.number().int().min(0).optional(),
});
export type SelectMultipleConfig = z.infer<typeof selectMultipleConfigSchema>;

const baseFormFieldSchema = z.object({
  id: z.string().min(1),
  label: z.string().trim().min(1, "Field label is required."),
  description: z.string().trim().optional(),
  required: z.boolean().default(false),
  order: z.number().int().default(1),
});

export const formFieldSchema = z.discriminatedUnion("type", [
  baseFormFieldSchema.extend({
    type: z.literal("SHORT_TEXT"),
    config: textConfigSchema.default({}),
  }),
  baseFormFieldSchema.extend({
    type: z.literal("LONG_TEXT"),
    config: textConfigSchema.default({}),
  }),
  baseFormFieldSchema.extend({
    type: z.literal("NUMBER"),
    config: numberConfigSchema.default({}),
  }),
  baseFormFieldSchema.extend({
    type: z.literal("SELECT_ONE"),
    config: selectOneConfigSchema.default({ options: [] }),
  }),
  baseFormFieldSchema.extend({
    type: z.literal("SELECT_MULTIPLE"),
    config: selectMultipleConfigSchema.default({ options: [] }),
  }),
]);

export type FormField = z.infer<typeof formFieldSchema>;
