"use client";

import React, { useState } from "react";
import { Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TournamentStatusBadge } from "@/features/tournaments/components/tournament-status-badge";
import { TournamentGeneralSection } from "./tournament-general-section";
import { TournamentGateSection } from "./tournament-gate-section";
import { TournamentTimelineSection } from "./tournament-timeline-section";
import { TournamentCrowdfundingSection } from "./tournament-crowdfunding-section";
import {
  EditTournamentSectionDialog,
  type SectionType,
} from "./edit-tournament-section-dialog";
import { ChangeTournamentStatusDialog } from "./change-tournament-status-dialog";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function TournamentHomeTab({
  tournamentId,
  summary,
  isOwner,
  onRefresh,
}: {
  tournamentId: string;
  summary: OrganizerSummary;
  isOwner: boolean;
  onRefresh: () => void;
}) {
  const { tournament, funding } = summary;
  const status = tournament.status;

  const [editSection, setEditSection] = useState<SectionType | null>(null);
  const [statusDialogOpen, setStatusDialogOpen] = useState(false);

  // Mutation rules from backend/README.md:
  // General Info: Editable in DRAFT, CROWDFUNDING, REGISTRATION, ONGOING. Locked in COMPLETED.
  const canEditGeneral = isOwner && status !== "COMPLETED";

  // Registration Gate: Editable in DRAFT, CROWDFUNDING. Locked in REGISTRATION, ONGOING, COMPLETED.
  const canEditGate = isOwner && (status === "DRAFT" || status === "CROWDFUNDING");

  // Registration Deadline & Timeline: Editable in DRAFT, CROWDFUNDING, REGISTRATION. Locked in ONGOING, COMPLETED.
  const canEditTimeline =
    isOwner && status !== "ONGOING" && status !== "COMPLETED";

  // Crowdfunding: Editable in DRAFT. Locked once CROWDFUNDING launches.
  const canEditCrowdfunding = isOwner && status === "DRAFT";

  const showCrowdfunding = funding.goalAmount > 0 || status === "DRAFT";

  return (
    <div className="flex flex-col gap-10">
      {/* Top Action Bar for Home Tab */}
      <div className="border-border bg-card/40 flex flex-wrap items-center justify-between gap-4 rounded-xl border p-4 sm:p-5">
        <div className="flex items-center gap-3">
          <span className="text-muted-foreground text-xs font-semibold uppercase tracking-wider">
            Status
          </span>
          <TournamentStatusBadge
            status={tournament.status}
            registrationDeadline={tournament.registrationDeadline}
          />
        </div>

        {isOwner && status !== "COMPLETED" ? (
          <Button
            size="sm"
            onClick={() => setStatusDialogOpen(true)}
            className="font-medium"
          >
            <Sparkles data-icon="inline-start" />
            Update Status
          </Button>
        ) : null}
      </div>

      {/* 1. General Information Section */}
      <TournamentGeneralSection
        summary={summary}
        canEdit={canEditGeneral}
        isOwner={isOwner}
        onEdit={() => setEditSection("general")}
      />

      {/* 2. Registration Gate Section */}
      <TournamentGateSection
        summary={summary}
        canEdit={canEditGate}
        onEdit={() => setEditSection("gate")}
      />

      {/* 3. Event Timeline Section */}
      <TournamentTimelineSection
        summary={summary}
        canEdit={canEditTimeline}
        onEdit={() => setEditSection("timeline")}
      />

      {/* 4. Crowdfunding Section (if active or draft) */}
      {showCrowdfunding ? (
        <TournamentCrowdfundingSection
          summary={summary}
          canEdit={canEditCrowdfunding}
          onEdit={() => setEditSection("crowdfunding")}
        />
      ) : null}

      {/* Modals */}
      <EditTournamentSectionDialog
        tournamentId={tournamentId}
        section={editSection}
        open={Boolean(editSection)}
        onOpenChange={(isOpen) => {
          if (!isOpen) setEditSection(null);
        }}
        summary={summary}
        onSuccess={onRefresh}
      />

      <ChangeTournamentStatusDialog
        tournamentId={tournamentId}
        currentStatus={tournament.status}
        funding={funding}
        open={statusDialogOpen}
        onOpenChange={setStatusDialogOpen}
        onSuccess={onRefresh}
      />
    </div>
  );
}
