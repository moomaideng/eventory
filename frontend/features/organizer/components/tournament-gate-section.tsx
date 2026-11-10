"use client";

import React from "react";
import { Lock, Pencil, Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  formatMoney,
  formatRegistrationType,
} from "@/features/tournaments/utils";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function TournamentGateSection({
  summary,
  canEdit,
  onEdit,
}: {
  summary: OrganizerSummary;
  canEdit: boolean;
  onEdit: () => void;
}) {
  const { tournament } = summary;

  const rosterLabel =
    tournament.registrationMode === "SOLO"
      ? "Solo"
      : tournament.minTeamSize === tournament.maxTeamSize
        ? `${tournament.minTeamSize} players per team`
        : `${tournament.minTeamSize} – ${tournament.maxTeamSize} players per team`;

  return (
    <section className="space-y-4">
      <div className="border-border flex items-center justify-between border-b pb-2.5">
        <div className="flex items-center gap-2">
          <Users className="text-primary size-5" />
          <h2 className="text-lg font-semibold tracking-tight">
            Registration Gate
          </h2>
        </div>
        {canEdit ? (
          <Button
            variant="ghost"
            size="sm"
            onClick={onEdit}
            aria-label="Edit Registration Gate"
            className="text-muted-foreground hover:text-foreground"
          >
            <Pencil data-icon="inline-start" />
            Edit
          </Button>
        ) : (
          <Badge variant="outline" className="text-muted-foreground gap-1">
            <Lock className="size-3" /> Locked
          </Badge>
        )}
      </div>

      <dl className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2 lg:grid-cols-4">
        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Registration Mode
          </dt>
          <dd className="mt-1 text-sm font-medium">
            {formatRegistrationType(tournament.registrationMode)}
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Roster Size
          </dt>
          <dd className="mt-1 text-sm font-medium">{rosterLabel}</dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Maximum Capacity
          </dt>
          <dd className="mt-1 text-sm font-medium">
            {tournament.capacity > 0
              ? `${tournament.capacity} ${
                  tournament.registrationMode === "SOLO" ? "players" : "teams"
                }`
              : "Not set"}
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Entry Fee
          </dt>
          <dd className="mt-1 text-sm font-medium">
            {formatMoney(tournament.entryFee, tournament.currency, true)}
          </dd>
        </div>
      </dl>
    </section>
  );
}
