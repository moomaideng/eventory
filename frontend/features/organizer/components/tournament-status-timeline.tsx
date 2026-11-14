"use client";

import { useState } from "react";
import { AlertCircle, ArrowRight, Flag, Sparkles, Trophy } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { extractProblemMessage } from "@/lib/schemas";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import type { DetailsField } from "@/features/organizer/tournament-details-form";
import {
  LifecycleConfirmDialog,
  type TransitionTarget,
} from "@/features/organizer/components/lifecycle-confirm-dialog";
import { LifecycleStepper } from "@/features/organizer/components/lifecycle-stepper";
import { LifecycleTransitionPanel } from "@/features/organizer/components/lifecycle-transition-panel";
import { useLifecycleData } from "@/features/organizer/hooks/use-lifecycle-data";
import {
  completionRequirements,
  crowdfundingRequirements,
  registrationRequirements,
  startRequirements,
  type RequirementFix,
  type WorkspaceTab,
} from "@/features/organizer/lifecycle";
import type { OrganizerSummary } from "@/features/organizer/utils";
import { formatMoney } from "@/features/tournaments/utils";

export function TournamentStatusTimeline({
  tournamentId,
  summary,
  isOwner,
  onSuccess,
  onEditField,
  onSelectTab,
}: {
  tournamentId: string;
  summary: OrganizerSummary;
  isOwner: boolean;
  onSuccess: () => void;
  onEditField: (field: DetailsField) => void;
  onSelectTab: (tab: WorkspaceTab) => void;
}) {
  const { authorizationHeader } = useAuth();
  const [confirmTarget, setConfirmTarget] = useState<TransitionTarget | null>(null);
  const [force, setForce] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

  const { tournament, funding } = summary;
  const status = tournament.status;
  const { now, tiers, matches } = useLifecycleData({ tournamentId, status, isOwner });

  const underGoal =
    status === "CROWDFUNDING" &&
    funding.raisedAmountMinor < funding.goalAmountMinor;
  const forceOpen = underGoal && force;

  function handleFix(fix: RequirementFix) {
    if (fix.kind === "field") onEditField(fix.field);
    else onSelectTab(fix.tab);
  }

  async function handleTransition(target: TransitionTarget) {
    if (!authorizationHeader || isSubmitting) return;
    setIsSubmitting(true);
    setApiError(null);
    try {
      const { error } = await apiClient.PATCH("/api/v1/tournaments/{id}/status", {
        params: { path: { id: tournamentId } },
        headers: { Authorization: authorizationHeader },
        body: {
          status: target,
          force: target === "REGISTRATION" && forceOpen ? true : undefined,
        },
      });
      if (error) {
        setApiError(extractProblemMessage(error, "Failed to update tournament status."));
        return;
      }
      onSuccess();
    } catch {
      setApiError("A network error occurred. Please try again.");
    } finally {
      setIsSubmitting(false);
      setConfirmTarget(null);
    }
  }

  const panelProps = {
    isOwner,
    submittingTarget: isSubmitting ? confirmTarget : null,
    onFix: handleFix,
    onRequest: setConfirmTarget,
  };

  return (
    <section
      aria-label="Tournament stage"
      className="flex flex-col gap-5 rounded-xl border bg-card p-4 sm:p-6"
    >
      <LifecycleStepper status={status} fundingGoal={funding.goalAmountMinor} />
      <Separator />

      {apiError ? (
        <Alert variant="destructive">
          <AlertCircle />
          <AlertTitle>Unable to update status</AlertTitle>
          <AlertDescription>{apiError}</AlertDescription>
        </Alert>
      ) : null}

      {status === "DRAFT" ? (
        <div className="flex flex-col gap-3">
          <div className="flex flex-col gap-1">
            <h3 className="text-base font-semibold">Next step: choose how to launch</h3>
            <p className="text-muted-foreground text-sm">
              Raise sponsor funding first, or skip straight to player registration.
            </p>
          </div>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <LifecycleTransitionPanel
              {...panelProps}
              layout="card"
              title="Crowdfunding first"
              description="Publish the tournament and sell sponsorship tiers before registration opens."
              items={crowdfundingRequirements(summary, tiers, now)}
              target="CROWDFUNDING"
              actionLabel="Open crowdfunding"
              actionVariant="outline"
              actionIcon={<Sparkles data-icon="inline-start" />}
            />
            <LifecycleTransitionPanel
              {...panelProps}
              layout="card"
              title="Direct registration"
              description="Skip crowdfunding and let players register right away."
              items={registrationRequirements(summary, now, false)}
              target="REGISTRATION"
              actionLabel="Open registration"
              actionIcon={<ArrowRight data-icon="inline-start" />}
            />
          </div>
        </div>
      ) : status === "CROWDFUNDING" ? (
        <LifecycleTransitionPanel
          {...panelProps}
          layout="row"
          title="Next step: open registration"
          description="End the crowdfunding campaign and let players register."
          items={registrationRequirements(summary, now, forceOpen)}
          target="REGISTRATION"
          actionLabel="Open registration"
          actionIcon={<ArrowRight data-icon="inline-start" />}
        >
          {underGoal && isOwner ? (
            <div className="flex items-start gap-2.5 rounded-lg border border-amber-500/20 bg-amber-500/10 px-3 py-2.5 text-sm text-amber-900 dark:text-amber-200">
              <input
                type="checkbox"
                id="force-reg-checkbox"
                checked={force}
                onChange={(e) => setForce(e.target.checked)}
                className="mt-0.5 size-4 shrink-0 rounded border-border accent-primary"
              />
              <label htmlFor="force-reg-checkbox" className="cursor-pointer">
                The goal isn&apos;t reached yet (
                {formatMoney(funding.remainingAmountMinor, funding.currency)} to
                go).
                Open registration anyway.
              </label>
            </div>
          ) : null}
        </LifecycleTransitionPanel>
      ) : status === "REGISTRATION" ? (
        <LifecycleTransitionPanel
          {...panelProps}
          layout="row"
          title="Next step: start the tournament"
          description="Close registration for good, lock rosters and begin bracket play."
          items={startRequirements(summary, matches, now)}
          target="ONGOING"
          actionLabel="Start tournament"
          actionIcon={<Flag data-icon="inline-start" />}
        />
      ) : status === "ONGOING" ? (
        <LifecycleTransitionPanel
          {...panelProps}
          layout="row"
          title="Next step: complete the tournament"
          description="Once every match has a result, archive the event as a final record."
          items={completionRequirements(matches)}
          target="COMPLETED"
          actionLabel="Complete tournament"
          actionIcon={<Trophy data-icon="inline-start" />}
        />
      ) : (
        <div className="flex items-start gap-3">
          <Trophy className="mt-0.5 size-5 shrink-0 text-primary" />
          <div className="flex flex-col gap-1">
            <h3 className="text-base font-semibold">Tournament concluded</h3>
            <p className="text-muted-foreground text-sm">
              Results are final and the tournament is read-only.
            </p>
          </div>
        </div>
      )}

      {!isOwner && status !== "COMPLETED" ? (
        <p className="text-muted-foreground text-sm">
          Only the tournament owner can change the stage.
        </p>
      ) : null}

      <LifecycleConfirmDialog
        target={confirmTarget}
        from={status}
        registrationDeadline={tournament.registrationDeadline}
        force={forceOpen}
        isSubmitting={isSubmitting}
        onCancel={() => setConfirmTarget(null)}
        onConfirm={(target) => void handleTransition(target)}
      />
    </section>
  );
}
