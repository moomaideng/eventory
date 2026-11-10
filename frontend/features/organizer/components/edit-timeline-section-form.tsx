"use client";

import React, { useState } from "react";
import { AlertCircle } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { extractProblemMessage } from "@/lib/schemas";
import { Button } from "@/components/ui/button";
import {
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { toLocalInput } from "@/features/organizer/utils";
import { editTimelineSectionSchema } from "../schemas";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function EditTimelineSectionForm({
  tournamentId,
  summary,
  onClose,
  onSuccess,
}: {
  tournamentId: string;
  summary: OrganizerSummary;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const { authorizationHeader } = useAuth();
  const { tournament } = summary;

  const [registrationDeadline, setRegistrationDeadline] = useState(
    toLocalInput(tournament.registrationDeadline)
  );
  const [startAt, setStartAt] = useState(toLocalInput(tournament.startAt));
  const [endAt, setEndAt] = useState(toLocalInput(tournament.endAt));

  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!authorizationHeader) return;
    setFieldErrors({});
    setApiError(null);

    const parsed = editTimelineSectionSchema.safeParse({
      registrationDeadline,
      startAt,
      endAt,
    });

    if (!parsed.success) {
      const flattened = parsed.error.flatten().fieldErrors;
      const mapped: Record<string, string> = {};
      for (const [key, msgs] of Object.entries(flattened)) {
        if (msgs?.[0]) mapped[key] = msgs[0];
      }
      setFieldErrors(mapped);
      return;
    }

    setIsSubmitting(true);
    try {
      const { error } = await apiClient.PATCH("/api/v1/tournaments/{tournamentId}", {
        params: { path: { tournamentId } },
        headers: { Authorization: authorizationHeader },
        body: {
          registrationDeadline: parsed.data.registrationDeadline
            ? new Date(parsed.data.registrationDeadline).toISOString()
            : undefined,
          startAt: parsed.data.startAt
            ? new Date(parsed.data.startAt).toISOString()
            : undefined,
          endAt: parsed.data.endAt
            ? new Date(parsed.data.endAt).toISOString()
            : undefined,
        },
      });

      if (error) {
        setApiError(
          extractProblemMessage(error, "Failed to update tournament schedule.")
        );
        return;
      }

      onClose();
      onSuccess();
    } catch {
      setApiError("Network failure. Please retry shortly.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <DialogHeader>
        <DialogTitle>Edit Timeline & Schedule</DialogTitle>
        <DialogDescription>
          Set key timestamps for registration deadline, tournament start, and conclusion.
        </DialogDescription>
      </DialogHeader>

      <div className="flex flex-col gap-4 py-4">
        {apiError ? (
          <div
            role="alert"
            className="bg-destructive/10 text-destructive border-destructive/20 flex items-start gap-2.5 rounded-lg border p-3 text-sm"
          >
            <AlertCircle className="mt-0.5 size-4 shrink-0" />
            <span>{apiError}</span>
          </div>
        ) : null}

        <FieldGroup>
          <Field data-invalid={Boolean(fieldErrors.registrationDeadline)}>
            <FieldLabel htmlFor="edit-regDeadline">
              Registration Deadline
            </FieldLabel>
            <Input
              id="edit-regDeadline"
              type="datetime-local"
              value={registrationDeadline}
              aria-invalid={Boolean(fieldErrors.registrationDeadline)}
              onChange={(e) => setRegistrationDeadline(e.target.value)}
            />
            {fieldErrors.registrationDeadline ? (
              <FieldError>{fieldErrors.registrationDeadline}</FieldError>
            ) : null}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.startAt)}>
            <FieldLabel htmlFor="edit-startAt">Tournament Start</FieldLabel>
            <Input
              id="edit-startAt"
              type="datetime-local"
              value={startAt}
              aria-invalid={Boolean(fieldErrors.startAt)}
              onChange={(e) => setStartAt(e.target.value)}
            />
            {fieldErrors.startAt ? (
              <FieldError>{fieldErrors.startAt}</FieldError>
            ) : null}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.endAt)}>
            <FieldLabel htmlFor="edit-endAt">Tournament End</FieldLabel>
            <Input
              id="edit-endAt"
              type="datetime-local"
              value={endAt}
              aria-invalid={Boolean(fieldErrors.endAt)}
              onChange={(e) => setEndAt(e.target.value)}
            />
            {fieldErrors.endAt ? (
              <FieldError>{fieldErrors.endAt}</FieldError>
            ) : null}
          </Field>
        </FieldGroup>
      </div>

      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          onClick={onClose}
          disabled={isSubmitting}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? (
            <>
              <Spinner data-icon="inline-start" />
              Saving...
            </>
          ) : (
            "Save Changes"
          )}
        </Button>
      </DialogFooter>
    </form>
  );
}
