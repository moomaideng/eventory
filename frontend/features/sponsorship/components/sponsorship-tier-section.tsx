"use client";

import { useMemo, useState } from "react";
import { AlertCircle, HandCoins, Pencil, Plus, Trash2 } from "lucide-react";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
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
import { DeleteTierDialog } from "@/features/sponsorship/components/delete-tier-dialog";
import { TierFormDialog } from "@/features/sponsorship/components/tier-form-dialog";
import { sortTiers, type SponsorshipTier } from "@/features/sponsorship/utils";
import { formatMoney } from "@/features/tournaments/utils";

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
  const { data, error, isLoading, refetch } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers",
    { params: { path: { tournamentId } } }
  );
  const tiers = useMemo(() => sortTiers(data?.items ?? []), [data?.items]);
  const close = () => setDialog({ kind: "closed" });

  return (
    <section className="flex min-w-0 flex-col gap-6" aria-labelledby="tiers-heading">
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
          <Button size="sm" onClick={() => setDialog({ kind: "form", tier: null })}>
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
            <span>Tiers are only shown for published tournaments.</span>
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
            <Card key={tier.id}>
              <CardHeader>
                <CardTitle className="wrap-anywhere">{tier.name}</CardTitle>
                <CardDescription>{formatMoney(tier.price, currency)}</CardDescription>
                <CardAction>
                  <Badge variant="secondary">
                    {tier.slots} {tier.slots === 1 ? "slot" : "slots"}
                  </Badge>
                </CardAction>
              </CardHeader>
              <CardContent>
                <ul className="text-muted-foreground flex list-disc flex-col gap-1 pl-5 text-sm">
                  {(tier.perks ?? []).map((perk) => (
                    <li key={perk} className="wrap-anywhere">
                      {perk}
                    </li>
                  ))}
                </ul>
              </CardContent>
              {editable ? (
                <CardFooter className="gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setDialog({ kind: "form", tier })}
                  >
                    <Pencil data-icon="inline-start" />
                    Edit
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setDialog({ kind: "delete", tier })}
                  >
                    <Trash2 data-icon="inline-start" />
                    Delete
                  </Button>
                </CardFooter>
              ) : null}
            </Card>
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
