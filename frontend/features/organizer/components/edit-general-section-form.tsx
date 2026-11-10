"use client";

import React, { useState } from "react";
import { AlertCircle, Globe } from "lucide-react";
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
import { Textarea } from "@/components/ui/textarea";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { Spinner } from "@/components/ui/spinner";
import { editGeneralSectionSchema } from "../schemas";
import type { OrganizerSummary } from "@/features/organizer/utils";

export function EditGeneralSectionForm({
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

  const [name, setName] = useState(tournament.name);
  const [game, setGame] = useState(tournament.game);
  const [location, setLocation] = useState(tournament.location);
  const [description, setDescription] = useState(tournament.description);
  const [detail, setDetail] = useState(tournament.detail || "");

  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!authorizationHeader) return;
    setFieldErrors({});
    setApiError(null);

    const parsed = editGeneralSectionSchema.safeParse({
      name,
      game,
      location,
      description,
      detail,
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
          name: parsed.data.name,
          game: parsed.data.game,
          location: parsed.data.location,
          description: parsed.data.description,
          detail: parsed.data.detail,
        },
      });

      if (error) {
        setApiError(
          extractProblemMessage(error, "Failed to update general information.")
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
        <DialogTitle>Edit General Information</DialogTitle>
        <DialogDescription>
          Update tournament title, game discipline, venue, short summary, and detailed rules.
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
          <Field data-invalid={Boolean(fieldErrors.name)}>
            <FieldLabel htmlFor="edit-name">Tournament Name</FieldLabel>
            <Input
              id="edit-name"
              value={name}
              maxLength={160}
              aria-invalid={Boolean(fieldErrors.name)}
              onChange={(e) => setName(e.target.value)}
            />
            {fieldErrors.name ? <FieldError>{fieldErrors.name}</FieldError> : null}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.game)}>
            <FieldLabel htmlFor="edit-game">Game Title</FieldLabel>
            <Input
              id="edit-game"
              value={game}
              maxLength={80}
              aria-invalid={Boolean(fieldErrors.game)}
              onChange={(e) => setGame(e.target.value)}
            />
            {fieldErrors.game ? <FieldError>{fieldErrors.game}</FieldError> : null}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.location)}>
            <div className="flex items-center justify-between">
              <FieldLabel htmlFor="edit-location">Location</FieldLabel>
              <Button
                type="button"
                variant="outline"
                size="xs"
                onClick={() => setLocation("Online")}
              >
                <Globe data-icon="inline-start" />
                Online
              </Button>
            </div>
            <Input
              id="edit-location"
              value={location}
              maxLength={160}
              aria-invalid={Boolean(fieldErrors.location)}
              onChange={(e) => setLocation(e.target.value)}
            />
            {fieldErrors.location ? (
              <FieldError>{fieldErrors.location}</FieldError>
            ) : null}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.description)}>
            <FieldLabel htmlFor="edit-description">Short Description</FieldLabel>
            <Textarea
              id="edit-description"
              value={description}
              maxLength={500}
              rows={3}
              aria-invalid={Boolean(fieldErrors.description)}
              onChange={(e) => setDescription(e.target.value)}
            />
            {fieldErrors.description ? (
              <FieldError>{fieldErrors.description}</FieldError>
            ) : null}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.detail)}>
            <div className="flex items-center justify-between">
              <FieldLabel htmlFor="edit-detail">
                Tournament Details & Rules
              </FieldLabel>
              <span className="text-[11px] text-muted-foreground">
                Markdown supported
              </span>
            </div>
            <MarkdownEditor
              id="edit-detail"
              value={detail}
              onChange={setDetail}
              rows={6}
              maxLength={10000}
            />
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
