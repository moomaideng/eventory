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
export const uuidSchema = z.string().uuid("Invalid identifier format");
