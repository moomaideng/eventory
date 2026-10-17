"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { TierCard } from "@/features/sponsorship/components/tier-card";
import { SponsorshipCheckoutDialog } from "@/features/sponsorship/components/sponsorship-checkout-dialog";
import { PublicSponsorsSection } from "@/features/sponsorship/components/public-sponsors-section";
import { sortTiers, type SponsorshipTier } from "@/features/sponsorship/utils";
import { useSponsorProfile } from "@/features/sponsor/hooks/use-sponsor-profile";
import { SponsorOnboardingModal } from "@/features/sponsor/components/sponsor-onboarding-modal";

/** Sponsorship packages on the public tournament page. Hidden when there are none. */
export function PublicTiersSection({
  tournamentId,
  currency,
  status,
  endsAt,
}: {
  tournamentId: string;
  currency: string;
  status: string;
  endsAt: string;
}) {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [selected, setSelected] = useState<SponsorshipTier | null>(null);
  const [beforeEnd] = useState(() => Date.now() < Date.parse(endsAt));
  const sponsor = useSponsorProfile({ enabled: Boolean(selected) });
  const profileFailed =
    Boolean(sponsor.error) &&
    !(
      typeof sponsor.error === "object" &&
      sponsor.error !== null &&
      "status" in sponsor.error &&
      sponsor.error.status === 404
    );
  const open =
    [
      "CROWDFUNDING",
      "REGISTRATION_OPEN",
      "REGISTRATION_CLOSED",
      "ONGOING",
    ].includes(status) && beforeEnd;
  const { data, isLoading, error, refetch } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers",
    { params: { path: { tournamentId } } },
    { staleTime: 30_000, retry: false }
  );
  const tiers = useMemo(() => sortTiers(data?.items ?? []), [data?.items]);

  if (isLoading) return <Skeleton className="h-48 rounded-xl" />;
  if (error)
    return (
      <Alert variant="destructive">
        <AlertTitle>Packages unavailable</AlertTitle>
        <AlertDescription>
          <Button variant="outline" onClick={() => void refetch()}>
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    );
  if (tiers.length === 0)
    return <PublicSponsorsSection tournamentId={tournamentId} />;

  function choose(tier: SponsorshipTier) {
    if (!authorizationHeader) {
      router.push(
        `/login?redirectTo=${encodeURIComponent(`/sponsor/tournaments/${tournamentId}`)}`
      );
      return;
    }
    setSelected(tier);
  }

  return (
    <>
      <section
        className="sponsorship-surface flex flex-col gap-4"
        aria-labelledby="public-tiers-heading"
      >
        <div className="flex flex-col gap-1">
          <h2
            id="public-tiers-heading"
            className="text-xl font-bold tracking-tight"
          >
            Sponsorship packages
          </h2>
          <p className="text-muted-foreground text-sm">
            {open
              ? "Back this tournament and get your brand in front of players and fans."
              : "Sponsorship checkout is closed for this tournament."}
          </p>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          {tiers.map((tier) => (
            <TierCard
              key={tier.id}
              tier={tier}
              currency={currency}
              footer={
                <Button
                  disabled={
                    !open ||
                    tier.availableSlots === 0 ||
                    authLoading ||
                    selected?.id === tier.id
                  }
                  onClick={() => choose(tier)}
                >
                  {tier.availableSlots === 0
                    ? "Sold out"
                    : selected?.id === tier.id
                      ? "Opening…"
                      : "Sponsor this tournament"}
                </Button>
              }
            />
          ))}
        </div>
        {authorizationHeader && (
          <p className="text-muted-foreground text-sm">
            Already started checkout?{" "}
            <Link
              href="/sponsor/payments"
              className="underline underline-offset-4"
            >
              Resume from My payments
            </Link>
            .
          </p>
        )}
        {selected && !sponsor.isLoading && profileFailed && (
          <Alert variant="destructive">
            <AlertTitle>Unable to load sponsor profile</AlertTitle>
            <AlertDescription>
              <Button variant="outline" onClick={() => void sponsor.refetch()}>
                Retry
              </Button>
              <Button variant="ghost" onClick={() => setSelected(null)}>
                Cancel
              </Button>
            </AlertDescription>
          </Alert>
        )}
      </section>
      <PublicSponsorsSection tournamentId={tournamentId} />
      {selected && sponsor.profile && (
        <SponsorshipCheckoutDialog
          key={selected.id}
          tournamentId={tournamentId}
          currency={currency}
          tier={selected}
          sponsorName={sponsor.profile.sponsorName}
          onClose={() => setSelected(null)}
        />
      )}
      {selected && sponsor.needsOnboarding && !profileFailed && (
        <SponsorOnboardingModal open onCancel={() => setSelected(null)} />
      )}
    </>
  );
}
