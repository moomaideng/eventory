"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Layers } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { TierCard } from "@/features/sponsorship/components/tier-card";
import { SponsorshipCheckoutDialog } from "@/features/sponsorship/components/sponsorship-checkout-dialog";
import { sortTiers, type SponsorshipTier } from "@/features/sponsorship/utils";
import { useSponsorProfile } from "@/features/sponsor/hooks/use-sponsor-profile";
import { SponsorOnboardingModal } from "@/features/sponsor/components/sponsor-onboarding-modal";

/** Sponsorship packages on the public tournament page. Hidden when there are none. */
export function PublicTiersSection({
  tournamentId,
  currency,
  status,
  endsAt,
  sponsorMode = false,
}: {
  tournamentId: string;
  currency: string;
  status: string;
  endsAt?: string;
  sponsorMode?: boolean;
}) {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const router = useRouter();
  const [selected, setSelected] = useState<SponsorshipTier | null>(null);
  const [beforeEnd] = useState(
    () => !endsAt || Date.now() < Date.parse(endsAt)
  );
  const sponsor = useSponsorProfile({
    enabled: sponsorMode && Boolean(selected),
  });
  const profileFailed =
    Boolean(sponsor.error) &&
    !(
      typeof sponsor.error === "object" &&
      sponsor.error !== null &&
      "status" in sponsor.error &&
      sponsor.error.status === 404
    );
  const open =
    ["CROWDFUNDING", "REGISTRATION", "ONGOING"].includes(status) && beforeEnd;
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
      <Card
        className="sponsorship-surface"
        aria-labelledby="public-tiers-heading"
      >
        <CardHeader>
          <CardTitle id="public-tiers-heading">Sponsorship packages</CardTitle>
          <CardDescription>
            {sponsorMode
              ? open
                ? "Back this tournament and get your brand in front of players and fans."
                : "Sponsorship checkout is closed for this tournament."
              : "Sponsorship packages offered by the organizer."}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          {tiers.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Layers />
                </EmptyMedia>
                <EmptyTitle>No sponsorship packages yet</EmptyTitle>
                <EmptyDescription>
                  Sponsorship tiers offered by the organizer will appear here.
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              {tiers.map((tier) => (
                <TierCard
                  key={tier.id}
                  tier={tier}
                  currency={currency}
                  className="bg-muted/30 shadow-none"
                  footer={
                    sponsorMode && (
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
                    )
                  }
                />
              ))}
            </div>
          )}
          {sponsorMode && authorizationHeader && (
            <p className="text-muted-foreground text-sm">
              Already started checkout?{" "}
              <Link
                href="/payments"
                className="underline underline-offset-4"
              >
                Resume from My payments
              </Link>
              .
            </p>
          )}
          {sponsorMode && selected && !sponsor.isLoading && profileFailed && (
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
        </CardContent>
      </Card>
      {sponsorMode && selected && sponsor.profile && (
        <SponsorshipCheckoutDialog
          key={selected.id}
          tournamentId={tournamentId}
          currency={currency}
          tier={selected}
          sponsorName={sponsor.profile.sponsorName}
          websiteUrl={sponsor.profile.websiteUrl}
          logoUrl={sponsor.profile.logoUrl}
          onClose={() => setSelected(null)}
        />
      )}
      {sponsorMode && selected && sponsor.needsOnboarding && !profileFailed && (
        <SponsorOnboardingModal open onCancel={() => setSelected(null)} />
      )}
    </>
  );
}
