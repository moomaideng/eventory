"use client";

import { CheckCircle2, CircleHelp, Info, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import type {
  Requirement,
  RequirementFix,
  RequirementState,
} from "@/features/organizer/lifecycle";

const ICONS: Record<
  Exclude<RequirementState, "pending">,
  { icon: typeof Info; className: string; label: string }
> = {
  met: {
    icon: CheckCircle2,
    className: "text-emerald-600 dark:text-emerald-400",
    label: "Done",
  },
  unmet: { icon: XCircle, className: "text-destructive", label: "Needs attention" },
  unknown: { icon: CircleHelp, className: "text-muted-foreground", label: "Not verified" },
  info: { icon: Info, className: "text-muted-foreground", label: "Optional" },
};

export function LifecycleChecklist({
  items,
  onFix,
}: {
  items: Requirement[];
  /** Omit to render read-only (no fix actions). */
  onFix?: (fix: RequirementFix) => void;
}) {
  return (
    <ul className="flex flex-col gap-2.5">
      {items.map((item) => {
        const visual = item.state === "pending" ? null : ICONS[item.state];
        const Icon = visual?.icon;
        const showFix =
          onFix && item.fix && (item.state === "unmet" || item.state === "info");

        return (
          <li key={item.id} className="flex items-start gap-2.5">
            <span className="mt-0.5 flex size-4 shrink-0 items-center justify-center">
              {Icon ? (
                <Icon
                  className={cn("size-4", visual.className)}
                  aria-label={visual.label}
                />
              ) : (
                <Spinner className="size-4 text-muted-foreground" aria-label="Checking" />
              )}
            </span>
            <div className="flex min-w-0 flex-1 flex-col gap-0.5">
              <span
                className={cn(
                  "text-sm",
                  item.state === "met" || item.state === "unmet"
                    ? "font-medium text-foreground"
                    : "text-muted-foreground"
                )}
              >
                {item.label}
              </span>
              {item.detail ? (
                <span className="text-muted-foreground text-sm">{item.detail}</span>
              ) : null}
            </div>
            {showFix && item.fix ? (
              <Button
                variant="link"
                size="sm"
                className="h-auto shrink-0 px-0"
                onClick={() => {
                  if (item.fix) onFix?.(item.fix);
                }}
              >
                {item.fix.kind === "tab" && item.fix.tab === "bracket"
                  ? "Open bracket"
                  : item.state === "info"
                    ? "Set up"
                    : "Fix"}
              </Button>
            ) : null}
          </li>
        );
      })}
    </ul>
  );
}
