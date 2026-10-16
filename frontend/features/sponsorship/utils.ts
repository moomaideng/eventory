import type { components } from "@/lib/api/schema";

export type SponsorshipTier = components["schemas"]["SponsorshipTierResponse"];

export function sortTiers(tiers: SponsorshipTier[]): SponsorshipTier[] {
  return [...tiers].sort(
    (a, b) => a.sortOrder - b.sortOrder || b.price - a.price
  );
}

export function getErrorMessage(error: unknown): string {
  if (
    error &&
    typeof error === "object" &&
    "detail" in error &&
    typeof (error as { detail: unknown }).detail === "string"
  ) {
    return (error as { detail: string }).detail;
  }
  return "Something went wrong. Please try again.";
}
