"use client";

import { useMemo } from "react";
import { $api } from "@/lib/api/client";
import { Skeleton } from "@/components/ui/skeleton";
import { TierCard } from "@/features/sponsorship/components/tier-card";
import { sortTiers } from "@/features/sponsorship/utils";

/** Sponsorship packages on the public tournament page. Hidden when there are none. */
export function PublicTiersSection({
  tournamentId,
  currency,
}: {
  tournamentId: string;
  currency: string;
}) {
  const { data, isLoading } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers",
    { params: { path: { tournamentId } } },
    { staleTime: 30_000 }
  );
  const tiers = useMemo(() => sortTiers(data?.items ?? []), [data?.items]);

  if (isLoading) return <Skeleton className="h-48 rounded-xl" />;
  if (tiers.length === 0) return null;

  return (
    <section className="flex flex-col gap-4" aria-labelledby="public-tiers-heading">
      <div className="flex flex-col gap-1">
        <h2 id="public-tiers-heading" className="text-xl font-bold tracking-tight">
          Sponsorship packages
        </h2>
        <p className="text-muted-foreground text-sm">
          Back this tournament and get your brand in front of players and fans.
        </p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        {tiers.map((tier) => (
          <TierCard key={tier.id} tier={tier} currency={currency} />
        ))}
      </div>
    </section>
  );
}
