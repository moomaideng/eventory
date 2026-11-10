"use client";

import type { ReactNode } from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { LifecycleChecklist } from "@/features/organizer/components/lifecycle-checklist";
import type { TransitionTarget } from "@/features/organizer/components/lifecycle-confirm-dialog";
import {
  blockingCount,
  isChecking,
  type Requirement,
  type RequirementFix,
} from "@/features/organizer/lifecycle";

export function LifecycleTransitionPanel({
  tournamentId,
  title,
  description,
  items,
  target,
  actionLabel,
  actionIcon,
  actionVariant = "default",
  layout,
  isOwner,
  submittingTarget,
  onFix,
  onRequest,
  children,
}: {
  tournamentId: string;
  title: string;
  description: string;
  items: Requirement[];
  target: TransitionTarget;
  actionLabel: string;
  actionIcon: ReactNode;
  actionVariant?: "default" | "outline";
  /** "row": single next step, action on the right. "card": one of several paths. */
  layout: "row" | "card";
  isOwner: boolean;
  /** Target currently being submitted, if any; disables every action. */
  submittingTarget: TransitionTarget | null;
  onFix: (fix: Exclude<RequirementFix, { kind: "bracket" }>) => void;
  onRequest: (target: TransitionTarget) => void;
  children?: ReactNode;
}) {
  const blocking = blockingCount(items);
  const checking = isChecking(items);
  const reason = checking
    ? "Checking requirements…"
    : blocking > 0
      ? `Resolve ${blocking} ${blocking === 1 ? "item" : "items"} to continue`
      : "Ready when you are";

  return (
    <div
      className={cn(
        "flex flex-col gap-4",
        layout === "row" && "md:flex-row md:items-start md:justify-between md:gap-8",
        layout === "card" && "justify-between rounded-xl border bg-card p-4 sm:p-5"
      )}
    >
      <div className="flex min-w-0 flex-1 flex-col gap-4">
        <div className="flex flex-col gap-1">
          <h3 className="text-base font-semibold">{title}</h3>
          <p className="text-muted-foreground text-sm">{description}</p>
        </div>
        <LifecycleChecklist
          tournamentId={tournamentId}
          items={items}
          onFix={isOwner ? onFix : undefined}
        />
        {children}
      </div>

      {isOwner ? (
        <div
          className={cn(
            "flex flex-col gap-2",
            layout === "row" && "md:w-60 md:shrink-0 md:items-end"
          )}
        >
          <Button
            variant={actionVariant}
            className="w-full"
            disabled={submittingTarget !== null || checking || blocking > 0}
            onClick={() => onRequest(target)}
          >
            {submittingTarget === target ? (
              <Spinner data-icon="inline-start" />
            ) : (
              actionIcon
            )}
            {actionLabel}
          </Button>
          <p
            className={cn(
              "text-sm md:text-right",
              blocking > 0 && !checking ? "text-destructive" : "text-muted-foreground",
              layout === "card" && "md:text-left"
            )}
          >
            {reason}
          </p>
        </div>
      ) : null}
    </div>
  );
}
