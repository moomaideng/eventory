import type { components } from "@/lib/api/schema";

export type SponsorshipTier = components["schemas"]["SponsorshipTierResponse"];

export function checkoutErrorFallback(status: number): string {
  return status === 413
    ? "Logo upload was rejected as too large. Choose a smaller file and retry."
    : `Checkout request failed (HTTP ${status}). Retry to resume the same checkout.`;
}

export function sortTiers(tiers: SponsorshipTier[]): SponsorshipTier[] {
  return [...tiers].sort(
    (a, b) => a.sortOrder - b.sortOrder || b.price - a.price
  );
}
