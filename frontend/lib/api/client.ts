import createClient from "openapi-fetch";
import type { paths } from "@/lib/api/schema";

/**
 * Type-safe OpenAPI client using openapi-fetch.
 */
export const apiClient = createClient<paths>({
  baseUrl: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
});
