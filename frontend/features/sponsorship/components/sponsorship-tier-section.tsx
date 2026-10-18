"use client";

import { useMemo, useState } from "react";
import { AlertCircle, Archive, HandCoins, Pencil, Plus } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { TierCard } from "@/features/sponsorship/components/tier-card";
import { DeleteTierDialog } from "@/features/sponsorship/components/delete-tier-dialog";
import { TierFormDialog } from "@/features/sponsorship/components/tier-form-dialog";
import { sortTiers, type SponsorshipTier } from "@/features/sponsorship/utils";

type DialogState =
  | { kind: "closed" }
  | { kind: "form"; tier: SponsorshipTier | null }
  | { kind: "delete"; tier: SponsorshipTier };

export function SponsorshipTierSection({
  tournamentId,
  currency,
  editable,
}: {
  tournamentId: string;
  currency: string;
  editable: boolean;
}) {
  const [dialog, setDialog] = useState<DialogState>({ kind: "closed" });
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const {
    data,
    error,
    isLoading: queryLoading,
    refetch,
  } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers/manage",
    {
      params: { path: { tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { enabled: Boolean(authorizationHeader), retry: false }
  );
  const isLoading = authLoading || !authorizationHeader || queryLoading;
  const tiers = useMemo(() => sortTiers(data?.items ?? []), [data?.items]);
  const close = () => setDialog({ kind: "closed" });

  return (
    <section
      className="flex min-w-0 flex-col gap-6"
      aria-labelledby="tiers-heading"
    >
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex min-w-0 flex-col gap-1">
          <h2 id="tiers-heading" className="text-xl font-bold tracking-tight">
            Sponsorship Tiers
          </h2>
          <p className="text-muted-foreground text-sm">
            Packages sponsors can buy to back this tournament.
          </p>
        </div>
        {editable && !isLoading ? (
          <Button
            size="sm"
            onClick={() => setDialog({ kind: "form", tier: null })}
          >
            <Plus data-icon="inline-start" />
            New tier
          </Button>
        ) : null}
      </div>

      {isLoading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Skeleton className="h-48 rounded-xl" />
          <Skeleton className="h-48 rounded-xl" />
        </div>
      ) : error ? (
        <Alert variant="destructive">
          <AlertCircle />
          <AlertTitle>Unable to load sponsorship tiers</AlertTitle>
          <AlertDescription className="flex items-center justify-between gap-4">
            <span>Something went wrong while loading the tiers.</span>
            <Button variant="outline" size="sm" onClick={() => void refetch()}>
              Retry
            </Button>
          </AlertDescription>
        </Alert>
      ) : tiers.length === 0 ? (
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <HandCoins />
            </EmptyMedia>
            <EmptyTitle>No sponsorship tiers yet</EmptyTitle>
            <EmptyDescription>
              {editable
                ? "Create a tier such as Gold or Silver so brands can sponsor you."
                : "Tiers can no longer be changed for this tournament."}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {tiers.map((tier) => (
            <TierCard
              key={tier.id}
              tier={tier}
              currency={currency}
              footer={
                editable ? (
                  <>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={tier.locked}
                      onClick={() => setDialog({ kind: "form", tier })}
                    >
                      <Pencil data-icon="inline-start" />
                      Edit
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      disabled={tier.locked}
                      onClick={() => setDialog({ kind: "delete", tier })}
                    >
                      <Archive data-icon="inline-start" />
                      Archive
                    </Button>
                    {tier.locked && (
                      <span className="text-muted-foreground text-xs">
                        Locked by a pending or confirmed sponsorship.
                      </span>
                    )}
                  </>
                ) : null
              }
            />
          ))}
        </div>
      )}

      {dialog.kind === "form" ? (
        <TierFormDialog
          key={dialog.tier?.id ?? "new"}
          tournamentId={tournamentId}
          currency={currency}
          tier={dialog.tier}
          open
          onOpenChange={(open) => (open ? null : close())}
        />
      ) : null}
      <DeleteTierDialog
        tournamentId={tournamentId}
        tier={dialog.kind === "delete" ? dialog.tier : null}
        open={dialog.kind === "delete"}
        onOpenChange={(open) => (open ? null : close())}
      />
    </section>
  );
}
