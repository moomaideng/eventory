"use client";

import React, { useState } from "react";
import { TournamentStatusTimeline } from "./tournament-status-timeline";
import { TournamentGeneralSection } from "./tournament-general-section";
import { TournamentGateSection } from "./tournament-gate-section";
import { TournamentTimelineSection } from "./tournament-timeline-section";
import { TournamentCrowdfundingSection } from "./tournament-crowdfunding-section";
import {
  EditTournamentSectionDialog,
  type SectionType,
} from "./edit-tournament-section-dialog";
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
      {/* Tournament Status Lifecycle Timeline */}
      <TournamentStatusTimeline
        tournamentId={tournamentId}
        summary={summary}
        isOwner={isOwner}
        onSuccess={onRefresh}
      />

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
    </div>
  );
}
