"use client";

import React from "react";
import { CalendarDays, Clock3, Lock, Pencil } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { formatTournamentDate } from "@/features/tournaments/utils";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function TournamentTimelineSection({
  summary,
  canEdit,
  onEdit,
}: {
  summary: OrganizerSummary;
  canEdit: boolean;
  onEdit: () => void;
}) {
  const { tournament } = summary;

  return (
    <section className="space-y-4">
      <div className="border-border flex items-center justify-between border-b pb-2.5">
        <div className="flex items-center gap-2">
          <CalendarDays className="text-primary size-5" />
          <h2 className="text-lg font-semibold tracking-tight">
            Timeline & Schedule
          </h2>
        </div>
        {canEdit ? (
          <Button
            variant="ghost"
            size="sm"
            onClick={onEdit}
            aria-label="Edit Timeline & Schedule"
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

      <dl className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-3">
        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Registration Closes
          </dt>
          <dd className="mt-1 flex items-center gap-1.5 text-sm font-medium">
            <Clock3 className="text-muted-foreground size-4 shrink-0" />
            <span>{formatTournamentDate(tournament.registrationDeadline)}</span>
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Tournament Starts
          </dt>
          <dd className="mt-1 flex items-center gap-1.5 text-sm font-medium">
            <CalendarDays className="text-muted-foreground size-4 shrink-0" />
            <span>{formatTournamentDate(tournament.startAt)}</span>
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Tournament Ends
          </dt>
          <dd className="mt-1 flex items-center gap-1.5 text-sm font-medium">
            <CalendarDays className="text-muted-foreground size-4 shrink-0" />
            <span>{formatTournamentDate(tournament.endAt)}</span>
          </dd>
        </div>
      </dl>
    </section>
  );
}
