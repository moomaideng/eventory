import { z } from "zod";

/**
 * RFC 7807 problem details object schema from the backend API.
 */
export const problemDetailsSchema = z.object({
  detail: z.string(),
});

export type ProblemDetails = z.infer<typeof problemDetailsSchema>;

/**
 * Safely extracts the error detail string from an API response or falls back to a default message.
 */
export function extractProblemMessage(
  error: unknown,
  fallback: string
): string {
  const parsed = problemDetailsSchema.safeParse(error);
  return parsed.success ? parsed.data.detail : fallback;
}

/**
 * Reusable UUID validator for entity IDs (e.g. tournaments, accounts).
 */
export const uuidSchema = z.uuid("Invalid identifier format");

// Optional contact phone. The API does not validate phone numbers, so this only
// rejects obvious junk (letters, symbols) and keeps real formats such as
// "+66 81 234 5678" and "081-234-5678".
const PHONE_PATTERN = /^\+?[0-9][0-9\s()-]{5,}$/;

/**
 * Optional phone number for role profiles. An empty string means "no phone"
 * and becomes undefined, so the request body simply omits the field.
 */
export const optionalPhoneSchema = z
  .string()
  .trim()
  .max(20, "Phone number must be 20 characters or fewer.")
  .refine(
    (value) => value === "" || PHONE_PATTERN.test(value),
    "Enter a valid phone number, e.g. +66 81 234 5678."
  )
  .transform((value) => (value === "" ? undefined : value));
