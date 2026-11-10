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
import { editCrowdfundingSectionSchema } from "../schemas";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function EditCrowdfundingSectionForm({
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
  const { funding } = summary;

  const [fundingGoal, setFundingGoal] = useState(funding.goalAmount || 0);
  const [fundingDeadline, setFundingDeadline] = useState(
    toLocalInput(funding.deadline)
  );

  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!authorizationHeader) return;
    setFieldErrors({});
    setApiError(null);

    const parsed = editCrowdfundingSectionSchema.safeParse({
      fundingGoal,
      fundingDeadline,
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
          fundingGoal: parsed.data.fundingGoal,
          fundingDeadline: new Date(parsed.data.fundingDeadline).toISOString(),
        },
      });

      if (error) {
        setApiError(
          extractProblemMessage(error, "Failed to update crowdfunding settings.")
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
        <DialogTitle>Edit Crowdfunding Settings</DialogTitle>
        <DialogDescription>
          Configure the funding goal and campaign deadline while in Draft state.
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
          <Field data-invalid={Boolean(fieldErrors.fundingGoal)}>
            <FieldLabel htmlFor="edit-fundingGoal">Funding Goal (THB)</FieldLabel>
            <Input
              id="edit-fundingGoal"
              type="number"
              min={1}
              value={fundingGoal}
              aria-invalid={Boolean(fieldErrors.fundingGoal)}
              onChange={(e) =>
                setFundingGoal(parseInt(e.target.value, 10) || 0)
              }
            />
            {fieldErrors.fundingGoal ? (
              <FieldError>{fieldErrors.fundingGoal}</FieldError>
            ) : null}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.fundingDeadline)}>
            <FieldLabel htmlFor="edit-fundingDeadline">
              Funding Deadline
            </FieldLabel>
            <Input
              id="edit-fundingDeadline"
              type="datetime-local"
              value={fundingDeadline}
              aria-invalid={Boolean(fieldErrors.fundingDeadline)}
              onChange={(e) => setFundingDeadline(e.target.value)}
            />
            {fieldErrors.fundingDeadline ? (
              <FieldError>{fieldErrors.fundingDeadline}</FieldError>
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
