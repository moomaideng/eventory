"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import { AlertCircle, Lock } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { extractProblemMessage } from "@/lib/schemas";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Spinner } from "@/components/ui/spinner";
import { cn } from "@/lib/utils";
import {
  CrowdfundingFields,
  GeneralFields,
  RegistrationFields,
  ScheduleFields,
} from "@/features/organizer/components/edit-tournament-fields";
import {
  DETAILS_SECTIONS,
  SECTION_FIELDS,
  SECTION_LOCK_REASON,
  buildDetailsSubmission,
  changedFields,
  initialDetailsValues,
  isSectionEditable,
  showsCrowdfunding,
  type DetailsErrors,
  type DetailsFormValues,
  type DetailsField,
} from "@/features/organizer/tournament-details-form";
import type { OrganizerSummary } from "@/features/organizer/utils";

export type DetailsSection =
  "general" | "registration" | "schedule" | "crowdfunding";

const SECTION_LABEL: Record<DetailsSection, string> = {
  general: "General",
  registration: "Registration",
  schedule: "Schedule",
  crowdfunding: "Crowdfunding",
};

const SECTION_LOCK_DETAIL: Record<DetailsSection, string> = {
  general: "Completed tournaments are kept as they were played.",
  registration:
    "Format, capacity, and entry fee stay fixed once players can register.",
  schedule: "Dates can't change after the tournament has started.",
  crowdfunding:
    "The funding goal and deadline are fixed once the campaign launches.",
};

type EditTournamentDialogProps = {
  tournamentId: string;
  summary: OrganizerSummary;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialField?: DetailsField;
  onSuccess: () => void;
};

export function EditTournamentDialog(props: EditTournamentDialogProps) {
  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className="flex max-h-[90vh] flex-col gap-0 overflow-hidden p-0 sm:max-w-2xl">
        {/* Mounted only while open, so form state re-initializes from summary on every open. */}
        <EditTournamentForm {...props} />
      </DialogContent>
    </Dialog>
  );
}

function EditTournamentForm({
  tournamentId,
  summary,
  onOpenChange,
  initialField = "name",
  onSuccess,
}: EditTournamentDialogProps) {
  const { authorizationHeader } = useAuth();
  const status = summary.tournament.status;
  const currency = summary.tournament.currency || "THB";
  const sections = showsCrowdfunding(summary)
    ? DETAILS_SECTIONS
    : DETAILS_SECTIONS.filter((section) => section !== "crowdfunding");

  const [initial] = useState(() => initialDetailsValues(summary));
  const [values, setValues] = useState(initial);
  const [errors, setErrors] = useState<DetailsErrors>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const sectionRefs = useRef<Partial<Record<DetailsSection, HTMLElement | null>>>(
    {}
  );
  const scrollTo = (section: DetailsSection) =>
    sectionRefs.current[section]?.scrollIntoView({ block: "start" });
  // Disabled (locked) fields can't take focus; fall back to their section.
  const focusField = (field: DetailsField) => {
    const section = sections.find((s) => SECTION_FIELDS[s].includes(field));
    const element = document.getElementById(`edit-${field}`);
    if (element && !element.hasAttribute("disabled")) {
      element.scrollIntoView({ block: "center" });
      element.focus({ preventScroll: true });
    } else if (section) {
      scrollTo(section);
    }
  };

  useEffect(() => {
    // Wait a tick so the dialog's own open-time autofocus doesn't override ours.
    const timer = window.setTimeout(() => focusField(initialField), 50);
    return () => window.clearTimeout(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const changed = changedFields(values, initial);
  const dirtyCount = changed.length;

  function handleChange(patch: Partial<DetailsFormValues>) {
    setValues((current) => ({ ...current, ...patch }));
    setErrors((current) => {
      const next = { ...current };
      for (const field of Object.keys(patch) as (keyof DetailsFormValues)[]) {
        delete next[field];
      }
      return next;
    });
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (!authorizationHeader || dirtyCount === 0) return;
    setApiError(null);

    const result = buildDetailsSubmission(values, initial, status);
    if (!result.ok) {
      setErrors(result.errors);
      const firstInvalid = sections
        .flatMap((section) => SECTION_FIELDS[section])
        .find((field) => result.errors[field]);
      if (firstInvalid) focusField(firstInvalid);
      return;
    }

    setErrors({});
    setIsSubmitting(true);
    try {
      const { error } = await apiClient.PATCH(
        "/api/v1/tournaments/{tournamentId}",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader },
          body: result.body,
        }
      );
      if (error) {
        setApiError(
          extractProblemMessage(error, "Failed to update tournament details.")
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

  function renderFields(section: DetailsSection) {
    const fieldProps = {
      values,
      errors,
      disabled: !isSectionEditable(section, status) || isSubmitting,
      onChange: handleChange,
    };
    switch (section) {
      case "general":
        return <GeneralFields {...fieldProps} />;
      case "registration":
        return <RegistrationFields {...fieldProps} currency={currency} />;
      case "schedule":
        return <ScheduleFields {...fieldProps} />;
      case "crowdfunding":
        return <CrowdfundingFields {...fieldProps} currency={currency} />;
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      noValidate
      className="flex min-h-0 flex-1 flex-col"
    >
      <DialogHeader className="px-6 pt-6 pb-4">
        <DialogTitle>Edit tournament details</DialogTitle>
        <DialogDescription>
          Changes apply to the public tournament page once saved.
        </DialogDescription>
      </DialogHeader>

      {apiError ? (
        <div className="px-6 pb-4">
          <Alert variant="destructive">
            <AlertCircle />
            <AlertTitle>Couldn&apos;t save changes</AlertTitle>
            <AlertDescription>{apiError}</AlertDescription>
          </Alert>
        </div>
      ) : null}

      <div className="min-h-0 flex-1 overflow-y-auto border-t px-6">
        {sections.map((section, index) => {
          const locked = !isSectionEditable(section, status);
          const dirty = SECTION_FIELDS[section].some((field) =>
            changed.includes(field)
          );
          return (
            <section
              key={section}
              ref={(node) => {
                sectionRefs.current[section] = node;
              }}
              aria-labelledby={`edit-section-${section}`}
              className={cn(
                "flex scroll-mt-0 flex-col gap-4 py-6",
                index > 0 && "border-t"
              )}
            >
              <div className="flex items-center gap-2">
                <h3
                  id={`edit-section-${section}`}
                  className="text-base font-semibold"
                >
                  {SECTION_LABEL[section]}
                </h3>
                {locked ? (
                  <Lock aria-label="Locked" className="text-muted-foreground size-3.5" />
                ) : null}
                {dirty ? (
                  <span className="text-primary text-xs font-medium">
                    Edited
                  </span>
                ) : null}
              </div>
              {locked ? (
                <Alert>
                  <Lock />
                  <AlertTitle>{SECTION_LOCK_REASON[section]}</AlertTitle>
                  <AlertDescription>
                    {SECTION_LOCK_DETAIL[section]}
                  </AlertDescription>
                </Alert>
              ) : null}
              {renderFields(section)}
            </section>
          );
        })}
      </div>

      <div className="bg-popover flex flex-col-reverse gap-3 border-t px-6 py-4 sm:flex-row sm:items-center sm:justify-between">
        <p className="text-muted-foreground text-sm">
          {dirtyCount === 0
            ? "No unsaved changes"
            : `${dirtyCount} unsaved ${dirtyCount === 1 ? "change" : "changes"}`}
        </p>
        <div className="flex flex-col-reverse gap-2 sm:flex-row">
          <DialogClose
            render={<Button type="button" variant="outline" />}
            disabled={isSubmitting}
          >
            Cancel
          </DialogClose>
          <Button type="submit" disabled={isSubmitting || dirtyCount === 0}>
            {isSubmitting ? <Spinner data-icon="inline-start" /> : null}
            {isSubmitting ? "Saving…" : "Save changes"}
          </Button>
        </div>
      </div>
    </form>
  );
}
