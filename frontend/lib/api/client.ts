import createClient from "openapi-fetch";
import type { paths } from "@/lib/api/schema";

/**
 * Type-safe OpenAPI client using openapi-fetch.
 */
const apiBaseUrl =
  typeof window === "undefined"
    ? process.env.INTERNAL_API_URL || process.env.NEXT_PUBLIC_API_URL
    : process.env.NEXT_PUBLIC_API_URL;

export const apiClient = createClient<paths>({
  baseUrl: apiBaseUrl || "http://localhost:8080",
});
