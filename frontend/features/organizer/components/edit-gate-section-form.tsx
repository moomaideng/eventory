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
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Spinner } from "@/components/ui/spinner";
import { editGateSectionSchema } from "../schemas";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function EditGateSectionForm({
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

  const [registrationMode, setRegistrationMode] = useState<"SOLO" | "TEAM">(
    tournament.registrationMode === "SOLO" ? "SOLO" : "TEAM"
  );
  const [minTeamSize, setMinTeamSize] = useState(tournament.minTeamSize || 1);
  const [maxTeamSize, setMaxTeamSize] = useState(tournament.maxTeamSize || 1);
  const [capacity, setCapacity] = useState(tournament.capacity || 16);
  const [entryFee, setEntryFee] = useState(tournament.entryFee || 0);

  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!authorizationHeader) return;
    setFieldErrors({});
    setApiError(null);

    const parsed = editGateSectionSchema.safeParse({
      registrationMode,
      minTeamSize: registrationMode === "SOLO" ? 1 : minTeamSize,
      maxTeamSize: registrationMode === "SOLO" ? 1 : maxTeamSize,
      capacity,
      entryFee,
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
          registrationMode: parsed.data.registrationMode,
          minTeamSize: parsed.data.minTeamSize,
          maxTeamSize: parsed.data.maxTeamSize,
          capacity: parsed.data.capacity,
          entryFee: parsed.data.entryFee,
        },
      });

      if (error) {
        setApiError(
          extractProblemMessage(error, "Failed to update registration gate.")
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
        <DialogTitle>Edit Registration Gate</DialogTitle>
        <DialogDescription>
          Configure roster limits, entry fees, registration mode, and participant capacity.
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
          <Field>
            <FieldLabel>Registration Mode</FieldLabel>
            <RadioGroup
              value={registrationMode}
              onValueChange={(val) => {
                const mode = val as "SOLO" | "TEAM";
                setRegistrationMode(mode);
                if (mode === "SOLO") {
                  setMinTeamSize(1);
                  setMaxTeamSize(1);
                }
              }}
              className="flex gap-4 pt-1"
            >
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <RadioGroupItem value="TEAM" />
                <span>Team</span>
              </label>
              <label className="flex items-center gap-2 text-sm cursor-pointer">
                <RadioGroupItem value="SOLO" />
                <span>Solo (1v1)</span>
              </label>
            </RadioGroup>
          </Field>

          {registrationMode === "TEAM" ? (
            <div className="grid grid-cols-2 gap-4">
              <Field data-invalid={Boolean(fieldErrors.minTeamSize)}>
                <FieldLabel htmlFor="edit-minTeamSize">Min Team Size</FieldLabel>
                <Input
                  id="edit-minTeamSize"
                  type="number"
                  min={1}
                  value={minTeamSize}
                  aria-invalid={Boolean(fieldErrors.minTeamSize)}
                  onChange={(e) =>
                    setMinTeamSize(parseInt(e.target.value, 10) || 1)
                  }
                />
                {fieldErrors.minTeamSize ? (
                  <FieldError>{fieldErrors.minTeamSize}</FieldError>
                ) : null}
              </Field>

              <Field data-invalid={Boolean(fieldErrors.maxTeamSize)}>
                <FieldLabel htmlFor="edit-maxTeamSize">Max Team Size</FieldLabel>
                <Input
                  id="edit-maxTeamSize"
                  type="number"
                  min={minTeamSize}
                  value={maxTeamSize}
                  aria-invalid={Boolean(fieldErrors.maxTeamSize)}
                  onChange={(e) =>
                    setMaxTeamSize(parseInt(e.target.value, 10) || 1)
                  }
                />
                {fieldErrors.maxTeamSize ? (
                  <FieldError>{fieldErrors.maxTeamSize}</FieldError>
                ) : null}
              </Field>
            </div>
          ) : null}

          <div className="grid grid-cols-2 gap-4">
            <Field data-invalid={Boolean(fieldErrors.capacity)}>
              <FieldLabel htmlFor="edit-capacity">Capacity</FieldLabel>
              <Input
                id="edit-capacity"
                type="number"
                min={1}
                value={capacity}
                aria-invalid={Boolean(fieldErrors.capacity)}
                onChange={(e) =>
                  setCapacity(parseInt(e.target.value, 10) || 1)
                }
              />
              {fieldErrors.capacity ? (
                <FieldError>{fieldErrors.capacity}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.entryFee)}>
              <FieldLabel htmlFor="edit-entryFee">Entry Fee (THB)</FieldLabel>
              <Input
                id="edit-entryFee"
                type="number"
                min={0}
                value={entryFee}
                aria-invalid={Boolean(fieldErrors.entryFee)}
                onChange={(e) =>
                  setEntryFee(parseInt(e.target.value, 10) || 0)
                }
              />
              {fieldErrors.entryFee ? (
                <FieldError>{fieldErrors.entryFee}</FieldError>
              ) : null}
            </Field>
          </div>
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
