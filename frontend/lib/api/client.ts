import createFetchClient from "openapi-fetch";
import createClient from "openapi-react-query";
import type { paths } from "@/lib/api/schema";

/**
 * Type-safe OpenAPI client using openapi-fetch.
 */
const apiBaseUrl =
  typeof window === "undefined"
    ? process.env.INTERNAL_API_URL || process.env.API_URL
    : process.env.API_URL;

export const apiClient = createFetchClient<paths>({
  baseUrl: apiBaseUrl || "http://localhost:8080",
});

export const createApiClient = createClient;
export const $api = createApiClient(apiClient);
