import type { components } from "@/lib/api/schema";

export type SponsorshipTier = components["schemas"]["SponsorshipTierResponse"];

export function sortTiers(tiers: SponsorshipTier[]): SponsorshipTier[] {
  return [...tiers].sort(
    (a, b) => a.sortOrder - b.sortOrder || b.price - a.price
  );
}
