"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Spinner } from "@/components/ui/spinner";
import type { LifecycleStatus } from "@/features/organizer/lifecycle";
import { formatTournamentDate } from "@/features/tournaments/utils";

export type TransitionTarget = Exclude<LifecycleStatus, "DRAFT">;

const COPY: Record<TransitionTarget, { title: string; action: string }> = {
  CROWDFUNDING: { title: "Open crowdfunding?", action: "Open crowdfunding" },
  REGISTRATION: { title: "Open registration?", action: "Open registration" },
  ONGOING: { title: "Start the tournament?", action: "Start tournament" },
  COMPLETED: { title: "Complete the tournament?", action: "Complete tournament" },
};

function consequences(
  target: TransitionTarget,
  from: string,
  registrationDeadline: string | undefined,
  force: boolean
) {
  switch (target) {
    case "CROWDFUNDING":
      return [
        "The tournament becomes public.",
        "The funding goal, funding deadline and sponsorship tiers lock.",
      ];
    case "REGISTRATION":
      return [
        ...(from === "DRAFT" ? ["The tournament becomes public."] : []),
        ...(force ? ["Crowdfunding ends before the funding goal is reached."] : []),
        "Registration rules and registration questions lock.",
        `Players can register until ${formatTournamentDate(registrationDeadline)}.`,
      ];
    case "ONGOING":
      return [
        "Team rosters and the bracket structure lock.",
        "Registration closes permanently.",
      ];
    case "COMPLETED":
      return ["The tournament becomes read-only."];
  }
}

export function LifecycleConfirmDialog({
  target,
  from,
  registrationDeadline,
  force,
  isSubmitting,
  onCancel,
  onConfirm,
}: {
  target: TransitionTarget | null;
  from: string;
  registrationDeadline?: string;
  force: boolean;
  isSubmitting: boolean;
  onCancel: () => void;
  onConfirm: (target: TransitionTarget) => void;
}) {
  return (
    <AlertDialog
      open={target !== null}
      onOpenChange={(open) => {
        if (!open && !isSubmitting) onCancel();
      }}
    >
      {target ? (
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{COPY[target].title}</AlertDialogTitle>
            <AlertDialogDescription render={<div />}>
              <ul className="flex list-disc flex-col gap-1 pl-5 text-left">
                {consequences(target, from, registrationDeadline, force).map(
                  (line) => (
                    <li key={line}>{line}</li>
                  )
                )}
              </ul>
              <p className="mt-3 font-medium text-foreground">
                This can&apos;t be undone.
              </p>
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isSubmitting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              disabled={isSubmitting}
              onClick={() => onConfirm(target)}
            >
              {isSubmitting ? <Spinner data-icon="inline-start" /> : null}
              {COPY[target].action}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      ) : null}
    </AlertDialog>
  );
}
