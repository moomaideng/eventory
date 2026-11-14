import type { components } from "@/lib/api/schema";

export type SponsorshipTier = components["schemas"]["SponsorshipTierResponse"];

export function checkoutErrorFallback(status: number): string {
  return `Checkout request failed (HTTP ${status}). Retry to resume the same checkout.`;
}

export function sortTiers(tiers: SponsorshipTier[]): SponsorshipTier[] {
  return [...tiers].sort(
    (a, b) => b.priceMinor - a.priceMinor || a.name.localeCompare(b.name)
  );
}
