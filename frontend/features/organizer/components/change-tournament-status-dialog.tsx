"use client";

import React, { useState } from "react";
import { AlertCircle, ArrowRight, CheckCircle2 } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { extractProblemMessage } from "@/lib/schemas";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogPortal,
  DialogOverlay,
  DialogTitle,
} from "@/components/ui/dialog";
import { Spinner } from "@/components/ui/spinner";
import { TournamentStatusBadge } from "@/features/tournaments/components/tournament-status-badge";
import type { OrganizerSummary } from "@/features/organizer/utils";

type TargetStatus = "DRAFT" | "CROWDFUNDING" | "REGISTRATION" | "ONGOING" | "COMPLETED";

interface StatusTransitionOption {
  targetStatus: TargetStatus;
  title: string;
  description: string;
  requirements: string[];
}

function getAvailableTransitions(currentStatus: string): StatusTransitionOption[] {
  switch (currentStatus) {
    case "DRAFT":
      return [
        {
          targetStatus: "CROWDFUNDING",
          title: "Launch Crowdfunding Campaign",
          description:
            "Opens the tournament to public sponsorship pledges. Freezes funding goal, deadline, and sponsor tiers once active.",
          requirements: [
            "Funding goal > 0 and future deadline",
            "At least one active sponsorship tier configured",
          ],
        },
        {
          targetStatus: "REGISTRATION",
          title: "Open Registration Directly",
          description:
            "Bypasses crowdfunding and opens player / team registration. Freezes capacity and roster rules once active.",
          requirements: [
            "Complete registration gate (capacity, fees, roster sizes)",
            "Complete timeline (Registration Deadline < Start Time < End Time)",
          ],
        },
      ];
    case "CROWDFUNDING":
      return [
        {
          targetStatus: "REGISTRATION",
          title: "Advance to Registration",
          description:
            "Concludes the crowdfunding campaign and opens competitor registration. Freezes registration gate rules.",
          requirements: [
            "Complete registration gate & timeline parameters",
            "Funding deadline must be before registration deadline",
            "Funding goal reached (or force open early)",
          ],
        },
      ];
    case "REGISTRATION":
      return [
        {
          targetStatus: "ONGOING",
          title: "Start Tournament (Ongoing)",
          description:
            "Closes registration and advances the tournament into live match play. Freezes bracket structure and schedules.",
          requirements: [
            "Registration deadline must have passed",
            "Zero pending / unresolved team applications",
            "At least 2 accepted teams",
            "Valid bracket structure with all Stage 1 slots seeded",
          ],
        },
      ];
    case "ONGOING":
      return [
        {
          targetStatus: "COMPLETED",
          title: "Complete Tournament",
          description:
            "Marks the competition complete and seals the tournament into a read-only historical archive.",
          requirements: [
            "All bracket matches must be finished",
            "Championship match must have a single champion recorded",
          ],
        },
      ];
    default:
      return [];
  }
}

export function ChangeTournamentStatusDialog({
  tournamentId,
  currentStatus,
  funding,
  open,
  onOpenChange,
  onSuccess,
}: {
  tournamentId: string;
  currentStatus: string;
  funding: OrganizerSummary["funding"];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}) {
  const { authorizationHeader } = useAuth();
  const [selectedTarget, setSelectedTarget] = useState<TargetStatus | null>(null);
  const [force, setForce] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const transitions = getAvailableTransitions(currentStatus);

  const isUnderFundingGoal =
    currentStatus === "CROWDFUNDING" &&
    funding.goalAmount > 0 &&
    funding.raisedAmount < funding.goalAmount;

  async function handleTransition(target: TargetStatus) {
    if (!authorizationHeader) return;
    setIsSubmitting(true);
    setApiError(null);

    try {
      const { error } = await apiClient.PATCH("/api/v1/tournaments/{id}/status", {
        params: { path: { id: tournamentId } },
        headers: { Authorization: authorizationHeader },
        body: {
          status: target,
          force: target === "REGISTRATION" && force,
        },
      });

      if (error) {
        setApiError(
          extractProblemMessage(
            error,
            "Could not advance tournament status. Please ensure all transition preconditions are satisfied."
          )
        );
        return;
      }

      onOpenChange(false);
      onSuccess();
    } catch {
      setApiError("Network failure. Please retry shortly.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPortal>
        <DialogOverlay />
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>Update Tournament Status</DialogTitle>
            <DialogDescription>
              Advance the tournament through its lifecycle state machine. Transitions enforce domain preconditions.
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-5 py-2">
            <div className="border-border bg-muted/30 flex items-center justify-between rounded-lg border p-3">
              <span className="text-muted-foreground text-xs font-medium uppercase">
                Current State
              </span>
              <TournamentStatusBadge status={currentStatus} />
            </div>

            {apiError ? (
              <div
                role="alert"
                className="bg-destructive/10 text-destructive border-destructive/20 flex items-start gap-2.5 rounded-lg border p-3 text-sm"
              >
                <AlertCircle className="mt-0.5 size-4 shrink-0" />
                <span>{apiError}</span>
              </div>
            ) : null}

            {transitions.length === 0 ? (
              <p className="text-muted-foreground py-4 text-center text-sm">
                This tournament is in a terminal status ({currentStatus}). No further status transitions are available.
              </p>
            ) : (
              <div className="flex flex-col gap-3">
                <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  Available Transitions
                </span>

                {transitions.map((option) => {
                  const isTargetSelected = selectedTarget === option.targetStatus;
                  return (
                    <div
                      key={option.targetStatus}
                      className={`border rounded-lg p-4 transition-all flex flex-col gap-3 ${
                        isTargetSelected
                          ? "border-primary bg-primary/5"
                          : "border-border hover:border-border/80"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <div className="flex items-center gap-2">
                            <h3 className="font-semibold text-sm">
                              {option.title}
                            </h3>
                            <ArrowRight className="size-3 text-muted-foreground" />
                            <TournamentStatusBadge status={option.targetStatus} />
                          </div>
                          <p className="text-muted-foreground text-xs mt-1">
                            {option.description}
                          </p>
                        </div>
                      </div>

                      <div className="border-border/60 bg-muted/20 rounded p-2.5 text-xs flex flex-col gap-1">
                        <span className="font-medium text-[11px] text-muted-foreground uppercase">
                          Preconditions required:
                        </span>
                        {option.requirements.map((req, i) => (
                          <div key={i} className="flex items-center gap-1.5 text-muted-foreground">
                            <CheckCircle2 className="size-3 text-primary shrink-0" />
                            <span>{req}</span>
                          </div>
                        ))}
                      </div>

                      {currentStatus === "CROWDFUNDING" &&
                      option.targetStatus === "REGISTRATION" &&
                      isUnderFundingGoal ? (
                        <div className="border-amber-500/20 bg-amber-500/10 text-amber-900 dark:text-amber-200 rounded p-2.5 text-xs flex items-center gap-2">
                          <input
                            type="checkbox"
                            id="force-start-checkbox"
                            checked={force}
                            onChange={(e) => setForce(e.target.checked)}
                            className="rounded"
                          />
                          <label htmlFor="force-start-checkbox" className="cursor-pointer">
                            Funding goal not reached. Force open registration early.
                          </label>
                        </div>
                      ) : null}

                      <Button
                        size="sm"
                        disabled={isSubmitting}
                        onClick={() => {
                          setSelectedTarget(option.targetStatus);
                          void handleTransition(option.targetStatus);
                        }}
                        className="self-end"
                      >
                        {isSubmitting && selectedTarget === option.targetStatus ? (
                          <>
                            <Spinner data-icon="inline-start" />
                            Transitioning...
                          </>
                        ) : (
                          `Transition to ${option.targetStatus}`
                        )}
                      </Button>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </DialogPortal>
    </Dialog>
  );
}
