"use client";

import React from "react";
import { Lock, MapPin, Pencil, Trophy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { MarkdownContent } from "@/components/ui/markdown-content";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function TournamentGeneralSection({
  summary,
  canEdit,
  isOwner,
  onEdit,
}: {
  summary: OrganizerSummary;
  canEdit: boolean;
  isOwner: boolean;
  onEdit: () => void;
}) {
  const { tournament } = summary;

  return (
    <section className="space-y-4">
      <div className="border-border flex items-center justify-between border-b pb-2.5">
        <div className="flex items-center gap-2">
          <Trophy className="text-primary size-5" />
          <h2 className="text-lg font-semibold tracking-tight">
            General Information
          </h2>
        </div>
        {canEdit ? (
          <Button
            variant="ghost"
            size="sm"
            onClick={onEdit}
            aria-label="Edit General Information"
            className="text-muted-foreground hover:text-foreground"
          >
            <Pencil data-icon="inline-start" />
            Edit
          </Button>
        ) : !isOwner ? null : (
          <Badge variant="outline" className="text-muted-foreground gap-1">
            <Lock className="size-3" /> Locked
          </Badge>
        )}
      </div>

      <dl className="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Tournament Name
          </dt>
          <dd className="mt-1 text-base font-semibold wrap-anywhere">
            {tournament.name}
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Game Title
          </dt>
          <dd className="mt-1 text-base font-medium">{tournament.game}</dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Location / Venue
          </dt>
          <dd className="mt-1 flex items-center gap-1.5 text-sm">
            <MapPin className="text-muted-foreground size-4 shrink-0" />
            <span>{tournament.location}</span>
          </dd>
        </div>

        <div>
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Host / Organization
          </dt>
          <dd className="mt-1 text-sm">{tournament.organizerName}</dd>
        </div>

        <div className="sm:col-span-2">
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Short Description
          </dt>
          <dd className="text-muted-foreground mt-1 text-sm leading-relaxed wrap-anywhere">
            {tournament.description || "No description provided."}
          </dd>
        </div>

        <div className="sm:col-span-2">
          <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
            Tournament Rules & Details
          </dt>
          <dd className="mt-2 text-sm">
            {tournament.detail ? (
              <div className="border-border/60 bg-muted/20 rounded-lg border p-4 leading-relaxed">
                <MarkdownContent content={tournament.detail} />
              </div>
            ) : (
              <p className="text-muted-foreground text-sm italic">
                No detailed rules or markdown instructions provided.
              </p>
            )}
          </dd>
        </div>
      </dl>
    </section>
  );
}
