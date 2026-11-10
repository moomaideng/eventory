"use client";

import React from "react";
import {
  Dialog,
  DialogContent,
  DialogPortal,
  DialogOverlay,
} from "@/components/ui/dialog";
import { EditGeneralSectionForm } from "./edit-general-section-form";
import { EditGateSectionForm } from "./edit-gate-section-form";
import { EditTimelineSectionForm } from "./edit-timeline-section-form";
import { EditCrowdfundingSectionForm } from "./edit-crowdfunding-section-form";
import type { OrganizerSummary } from "@/features/organizer/utils";

export type SectionType = "general" | "gate" | "timeline" | "crowdfunding";

export function EditTournamentSectionDialog({
  tournamentId,
  section,
  open,
  onOpenChange,
  summary,
  onSuccess,
}: {
  tournamentId: string;
  section: SectionType | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  summary: OrganizerSummary;
  onSuccess: () => void;
}) {
  if (!section) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPortal>
        <DialogOverlay />
        <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
          {section === "general" ? (
            <EditGeneralSectionForm
              key="general"
              tournamentId={tournamentId}
              summary={summary}
              onClose={() => onOpenChange(false)}
              onSuccess={onSuccess}
            />
          ) : section === "gate" ? (
            <EditGateSectionForm
              key="gate"
              tournamentId={tournamentId}
              summary={summary}
              onClose={() => onOpenChange(false)}
              onSuccess={onSuccess}
            />
          ) : section === "timeline" ? (
            <EditTimelineSectionForm
              key="timeline"
              tournamentId={tournamentId}
              summary={summary}
              onClose={() => onOpenChange(false)}
              onSuccess={onSuccess}
            />
          ) : section === "crowdfunding" ? (
            <EditCrowdfundingSectionForm
              key="crowdfunding"
              tournamentId={tournamentId}
              summary={summary}
              onClose={() => onOpenChange(false)}
              onSuccess={onSuccess}
            />
          ) : null}
        </DialogContent>
      </DialogPortal>
    </Dialog>
  );
}
