import React from "react";
import { Check } from "lucide-react";
import { cn } from "@/lib/utils";

const STEPS = [
  { id: "DRAFT", label: "Draft" },
  { id: "CROWDFUNDING", label: "Crowdfunding" },
  { id: "REGISTRATION", label: "Registration" },
  { id: "ONGOING", label: "Ongoing" },
  { id: "COMPLETED", label: "Completed" },
] as const;

type StepState = "done" | "current" | "upcoming" | "skipped";

export function LifecycleStepper({
  status,
  fundingGoal,
}: {
  status: string;
  fundingGoal: number;
}) {
  const currentIndex = STEPS.findIndex((step) => step.id === status);
  const isCompleted = status === "COMPLETED";

  const states: StepState[] = STEPS.map((step, idx) => {
    if (idx === currentIndex && !isCompleted) return "current";
    if (step.id === "CROWDFUNDING" && currentIndex > idx && fundingGoal === 0) {
      return "skipped";
    }
    return idx <= currentIndex ? "done" : "upcoming";
  });

  return (
    <ol className="flex w-full items-start" aria-label="Tournament stages">
      {STEPS.map((step, idx) => {
        const state = states[idx];
        return (
          <React.Fragment key={step.id}>
            <li
              className="flex min-w-0 flex-col items-center gap-1.5"
              aria-current={state === "current" ? "step" : undefined}
            >
              <span
                className={cn(
                  "flex size-8 shrink-0 items-center justify-center rounded-full text-sm font-semibold",
                  state === "current" &&
                    "bg-primary text-primary-foreground ring-4 ring-primary/20",
                  state === "done" &&
                    "border border-primary/40 bg-primary/15 text-primary",
                  state === "upcoming" &&
                    "border border-border bg-muted/60 text-muted-foreground",
                  state === "skipped" &&
                    "border border-dashed border-muted-foreground/50 text-muted-foreground"
                )}
              >
                {state === "done" ? <Check className="size-4" /> : idx + 1}
              </span>
              <span
                className={cn(
                  "text-center text-xs sm:text-sm",
                  state === "current" && "font-semibold text-primary",
                  state === "done" && "font-medium text-foreground",
                  (state === "upcoming" || state === "skipped") &&
                    "text-muted-foreground"
                )}
              >
                {step.label}
              </span>
              {state === "skipped" ? (
                <span className="text-muted-foreground text-xs">Skipped</span>
              ) : state === "current" ? (
                <span className="sr-only">(current stage)</span>
              ) : null}
            </li>
            {idx < STEPS.length - 1 ? (
              <li
                aria-hidden
                className={cn(
                  "mx-1 mt-4 h-0 flex-1 border-t-2 sm:mx-3",
                  idx < currentIndex ? "border-primary" : "border-border",
                  (states[idx] === "skipped" || states[idx + 1] === "skipped") &&
                    "border-dashed border-muted-foreground/40"
                )}
              />
            ) : null}
          </React.Fragment>
        );
      })}
    </ol>
  );
}
