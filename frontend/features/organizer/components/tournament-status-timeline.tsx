"use client";

import React, { useState } from "react";
import { AlertCircle, ArrowRight, Check, CheckCircle2, Sparkles } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { extractProblemMessage } from "@/lib/schemas";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Spinner } from "@/components/ui/spinner";
import type { OrganizerSummary } from "@/features/organizer/utils";

type TargetStatus =
  | "DRAFT"
  | "CROWDFUNDING"
  | "REGISTRATION"
  | "ONGOING"
  | "COMPLETED";

const STATUS_STEPS = [
  { id: "DRAFT", label: "Draft", step: 1 },
  { id: "CROWDFUNDING", label: "Crowdfunding", step: 2 },
  { id: "REGISTRATION", label: "Registration", step: 3 },
  { id: "ONGOING", label: "Ongoing", step: 4 },
  { id: "COMPLETED", label: "Completed", step: 5 },
] as const;

export function TournamentStatusTimeline({
  tournamentId,
  summary,
  isOwner,
  onSuccess,
}: {
  tournamentId: string;
  summary: OrganizerSummary;
  isOwner: boolean;
  onSuccess: () => void;
}) {
  const { authorizationHeader } = useAuth();
  const [targetStatus, setTargetStatus] = useState<TargetStatus | null>(null);
  const [force, setForce] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const { tournament, funding } = summary;
  const currentStatus = tournament.status;

  const currentStepIndex = STATUS_STEPS.findIndex(
    (s) => s.id === currentStatus
  );

  const isUnderFundingGoal =
    currentStatus === "CROWDFUNDING" &&
    funding.goalAmount > 0 &&
    funding.raisedAmount < funding.goalAmount;

  async function handleTransition(target: TargetStatus) {
    if (!authorizationHeader || isSubmitting) return;
    setIsSubmitting(true);
    setTargetStatus(target);
    setApiError(null);

    try {
      const { error } = await apiClient.PATCH(
        "/api/v1/tournaments/{id}/status",
        {
          params: { path: { id: tournamentId } },
          headers: { Authorization: authorizationHeader },
          body: {
            status: target,
            force: target === "REGISTRATION" && force ? true : undefined,
          },
        }
      );

      if (error) {
        setApiError(
          extractProblemMessage(error, "Failed to update tournament status.")
        );
        return;
      }

      onSuccess();
    } catch {
      setApiError("A network error occurred. Please try again.");
    } finally {
      setIsSubmitting(false);
      setTargetStatus(null);
    }
  }

  return (
    <div className="border-border bg-card/40 flex flex-col gap-4 rounded-xl border p-4 sm:p-5">
      {/* 1. Centered Timeline */}
      <div className="w-full max-w-2xl mx-auto px-1 sm:px-4">
        <div className="flex items-center justify-between">
          {STATUS_STEPS.map((step, idx) => {
            const isCompleted = idx < currentStepIndex;
            const isCurrent = idx === currentStepIndex;

            return (
              <React.Fragment key={step.id}>
                <div className="flex flex-col items-center">
                  <div
                    className={cn(
                      "size-7 sm:size-8 rounded-full flex items-center justify-center text-xs font-semibold transition-all",
                      isCurrent
                        ? "bg-primary text-primary-foreground ring-4 ring-primary/20 shadow-xs"
                        : isCompleted
                        ? "bg-primary/20 text-primary border border-primary/40"
                        : "bg-muted/60 text-muted-foreground border border-border"
                    )}
                  >
                    {isCompleted ? (
                      <Check className="size-3.5 sm:size-4 stroke-[2.5]" />
                    ) : (
                      <span>{step.step}</span>
                    )}
                  </div>
                  <span
                    className={cn(
                      "mt-1.5 text-[10px] sm:text-xs tracking-tight text-center whitespace-nowrap",
                      isCurrent
                        ? "text-primary font-bold"
                        : isCompleted
                        ? "text-foreground font-medium"
                        : "text-muted-foreground"
                    )}
                  >
                    {step.label}
                  </span>
                </div>

                {idx < STATUS_STEPS.length - 1 ? (
                  <div
                    className={cn(
                      "flex-1 mx-1.5 sm:mx-3 -mt-5 h-0.5 border-t-2 border-dashed transition-colors",
                      idx < currentStepIndex
                        ? "border-primary"
                        : "border-border"
                    )}
                  />
                ) : null}
              </React.Fragment>
            );
          })}
        </div>
      </div>

      {/* 2. Error message if transition failed */}
      {apiError ? (
        <Alert variant="destructive" className="py-2.5">
          <AlertCircle className="size-4" />
          <AlertTitle className="text-xs font-semibold">
            Unable to update status
          </AlertTitle>
          <AlertDescription className="text-xs mt-0.5">
            {apiError}
          </AlertDescription>
        </Alert>
      ) : null}

      {/* 3. Action options & requirements for next status */}
      <div className="border-t border-border/50 pt-3.5">
        {currentStatus === "DRAFT" ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {/* Option A: Open for Crowdfunding */}
            <div className="flex flex-col justify-between rounded-lg border border-border/70 bg-card/60 p-3.5 gap-3">
              <div>
                <span className="text-xs font-semibold text-foreground">
                  Option 1: Crowdfunding Campaign
                </span>
                <p className="text-muted-foreground text-[11px] mt-0.5">
                  Open for public crowdfunding and sponsor tier purchases before
                  player registration.
                </p>
                <div className="mt-2.5 flex flex-col gap-1 text-[11px] text-muted-foreground">
                  <span className="font-semibold text-[10px] uppercase tracking-wider text-muted-foreground/80">
                    Requirements:
                  </span>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="size-3 text-primary shrink-0" />
                    <span>Funding goal &gt; 0 and future funding deadline</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="size-3 text-primary shrink-0" />
                    <span>At least 1 active sponsorship tier configured</span>
                  </div>
                </div>
              </div>

              {isOwner ? (
                <Button
                  size="sm"
                  variant="outline"
                  className="w-full font-medium"
                  disabled={isSubmitting}
                  onClick={() => void handleTransition("CROWDFUNDING")}
                >
                  {isSubmitting && targetStatus === "CROWDFUNDING" ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <Sparkles data-icon="inline-start" />
                  )}
                  Open for crowdfunding
                </Button>
              ) : null}
            </div>

            {/* Option B: Open for Registration */}
            <div className="flex flex-col justify-between rounded-lg border border-border/70 bg-card/60 p-3.5 gap-3">
              <div>
                <span className="text-xs font-semibold text-foreground">
                  Option 2: Direct Registration
                </span>
                <p className="text-muted-foreground text-[11px] mt-0.5">
                  Bypass crowdfunding and open registration for teams and players
                  directly.
                </p>
                <div className="mt-2.5 flex flex-col gap-1 text-[11px] text-muted-foreground">
                  <span className="font-semibold text-[10px] uppercase tracking-wider text-muted-foreground/80">
                    Requirements:
                  </span>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="size-3 text-primary shrink-0" />
                    <span>
                      Registration gate (capacity &ge; 1, fee &ge; 0, team size)
                    </span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <CheckCircle2 className="size-3 text-primary shrink-0" />
                    <span>
                      Timeline (now &lt; deadline &lt; startsAt &lt; endsAt)
                    </span>
                  </div>
                </div>
              </div>

              {isOwner ? (
                <Button
                  size="sm"
                  className="w-full font-medium"
                  disabled={isSubmitting}
                  onClick={() => void handleTransition("REGISTRATION")}
                >
                  {isSubmitting && targetStatus === "REGISTRATION" ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <ArrowRight data-icon="inline-end" />
                  )}
                  Open for registration
                </Button>
              ) : null}
            </div>
          </div>
        ) : currentStatus === "CROWDFUNDING" ? (
          <div className="flex flex-col justify-between rounded-lg border border-border/70 bg-card/60 p-3.5 gap-3 max-w-xl mx-auto w-full">
            <div>
              <span className="text-xs font-semibold text-foreground">
                Advance to Registration
              </span>
              <p className="text-muted-foreground text-[11px] mt-0.5">
                Conclude crowdfunding campaign and open competitor registration.
              </p>
              <div className="mt-2.5 flex flex-col gap-1 text-[11px] text-muted-foreground">
                <span className="font-semibold text-[10px] uppercase tracking-wider text-muted-foreground/80">
                  Requirements:
                </span>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>Registration gate & timeline complete</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>
                    Funding deadline must be before registration deadline
                  </span>
                </div>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>Funding goal reached (or force open early)</span>
                </div>
              </div>

              {isUnderFundingGoal && isOwner ? (
                <div className="mt-3 flex items-center gap-2 rounded-md border border-amber-500/20 bg-amber-500/10 px-2.5 py-1.5 text-xs text-amber-900 dark:text-amber-200">
                  <input
                    type="checkbox"
                    id="force-reg-checkbox"
                    checked={force}
                    onChange={(e) => setForce(e.target.checked)}
                    className="rounded border-border text-primary focus:ring-primary"
                  />
                  <label
                    htmlFor="force-reg-checkbox"
                    className="cursor-pointer text-[11px]"
                  >
                    Funding goal not yet reached. Force open registration early.
                  </label>
                </div>
              ) : null}
            </div>

            {isOwner ? (
              <Button
                size="sm"
                className="self-end font-medium"
                disabled={isSubmitting}
                onClick={() => void handleTransition("REGISTRATION")}
              >
                {isSubmitting && targetStatus === "REGISTRATION" ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <ArrowRight data-icon="inline-end" />
                )}
                Open for registration
              </Button>
            ) : null}
          </div>
        ) : currentStatus === "REGISTRATION" ? (
          <div className="flex flex-col justify-between rounded-lg border border-border/70 bg-card/60 p-3.5 gap-3 max-w-xl mx-auto w-full">
            <div>
              <span className="text-xs font-semibold text-foreground">
                Start Tournament (Ongoing)
              </span>
              <p className="text-muted-foreground text-[11px] mt-0.5">
                Close registration, finalize participant rosters, and start
                live bracket matches.
              </p>
              <div className="mt-2.5 flex flex-col gap-1 text-[11px] text-muted-foreground">
                <span className="font-semibold text-[10px] uppercase tracking-wider text-muted-foreground/80">
                  Requirements:
                </span>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>Registration deadline must have passed</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>All team applications resolved (0 pending)</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>At least 2 accepted teams</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>Bracket created and Stage 1 matches seeded</span>
                </div>
              </div>
            </div>

            {isOwner ? (
              <Button
                size="sm"
                className="self-end font-medium"
                disabled={isSubmitting}
                onClick={() => void handleTransition("ONGOING")}
              >
                {isSubmitting && targetStatus === "ONGOING" ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <ArrowRight data-icon="inline-end" />
                )}
                Start tournament
              </Button>
            ) : null}
          </div>
        ) : currentStatus === "ONGOING" ? (
          <div className="flex flex-col justify-between rounded-lg border border-border/70 bg-card/60 p-3.5 gap-3 max-w-xl mx-auto w-full">
            <div>
              <span className="text-xs font-semibold text-foreground">
                Complete Tournament
              </span>
              <p className="text-muted-foreground text-[11px] mt-0.5">
                Conclude all tournament competition and archive the event as a
                historical record.
              </p>
              <div className="mt-2.5 flex flex-col gap-1 text-[11px] text-muted-foreground">
                <span className="font-semibold text-[10px] uppercase tracking-wider text-muted-foreground/80">
                  Requirements:
                </span>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>All bracket matches completed</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <CheckCircle2 className="size-3 text-primary shrink-0" />
                  <span>Single undisputed champion recorded</span>
                </div>
              </div>
            </div>

            {isOwner ? (
              <Button
                size="sm"
                className="self-end font-medium"
                disabled={isSubmitting}
                onClick={() => void handleTransition("COMPLETED")}
              >
                {isSubmitting && targetStatus === "COMPLETED" ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <Check className="size-4" data-icon="inline-start" />
                )}
                Complete tournament
              </Button>
            ) : null}
          </div>
        ) : (
          <div className="rounded-lg border border-border/70 bg-card/60 p-3.5 text-center max-w-xl mx-auto w-full">
            <span className="text-xs font-semibold text-foreground">
              Tournament Concluded
            </span>
            <p className="text-muted-foreground text-[11px] mt-0.5">
              This tournament is completed and archived. Match records and results
              are finalized.
            </p>
          </div>
        )}

        {!isOwner && currentStatus !== "COMPLETED" ? (
          <p className="text-[11px] text-muted-foreground text-center mt-2.5">
            Only the tournament owner can transition between lifecycle stages.
          </p>
        ) : null}
      </div>
    </div>
  );
}
