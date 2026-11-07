"use client";

import { useState, type FormEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Pencil } from "lucide-react";
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { toLocalInput } from "@/features/organizer/utils";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { plannedEventDetailsSchema } from "../schemas";

type Details = {
  location: string;
  description: string;
  detail?: string;
  registrationDeadline: string;
  startAt: string;
  endAt: string;
};

export function PlannedEventDetailsDialog({
  tournamentId,
  details,
  fundingDeadline,
}: {
  tournamentId: string;
  details: Details;
  fundingDeadline: string;
}) {
  const [open, setOpen] = useState(false);
  const [location, setLocation] = useState(details.location);
  const [description, setDescription] = useState(details.description);
  const [detail, setDetail] = useState(details.detail || "");
  const [registrationDeadline, setRegistrationDeadline] = useState(
    toLocalInput(details.registrationDeadline)
  );
  const [startAt, setStartAt] = useState(toLocalInput(details.startAt));
  const [endAt, setEndAt] = useState(toLocalInput(details.endAt));
  const [errors, setErrors] = useState<Partial<Record<keyof Details, string>>>(
    {}
  );
  const [apiError, setApiError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  function onOpenChange(next: boolean) {
    if (next) {
      setLocation(details.location);
      setDescription(details.description);
      setDetail(details.detail || "");
      setRegistrationDeadline(toLocalInput(details.registrationDeadline));
      setStartAt(toLocalInput(details.startAt));
      setEndAt(toLocalInput(details.endAt));
      setErrors({});
      setApiError(null);
    }
    setOpen(next);
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const parsed = plannedEventDetailsSchema(fundingDeadline).safeParse({
      location,
      description,
      detail,
      registrationDeadline,
      startAt,
      endAt,
    });
    if (!parsed.success) {
      const fields = parsed.error.flatten().fieldErrors;
      setErrors({
        location: fields.location?.[0],
        description: fields.description?.[0],
        detail: fields.detail?.[0],
        registrationDeadline: fields.registrationDeadline?.[0],
        startAt: fields.startAt?.[0],
        endAt: fields.endAt?.[0],
      });
      return;
    }
    if (!authorizationHeader) {
      setApiError("Please sign in again.");
      return;
    }
    setErrors({});
    setApiError(null);
    setSaving(true);
    try {
      const { error } = await apiClient.PATCH(
        "/api/v1/tournaments/{tournamentId}",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader },
          body: {
            location: parsed.data.location,
            description: parsed.data.description,
            detail: parsed.data.detail,
            registrationDeadline: new Date(
              parsed.data.registrationDeadline
            ).toISOString(),
            startAt: new Date(parsed.data.startAt).toISOString(),
            endAt: new Date(parsed.data.endAt).toISOString(),
          },
        }
      );
      if (error) {
        setApiError(
          extractProblemMessage(error, "Could not update the event plan.")
        );
        return;
      }
      await queryClient.invalidateQueries();
      setOpen(false);
    } catch {
      setApiError("Connection failure. Please try again.");
    } finally {
      setSaving(false);
    }
  }

  const fields = [
    {
      key: "registrationDeadline" as const,
      label: "Planned registration deadline",
      value: registrationDeadline,
      onChange: setRegistrationDeadline,
    },
    {
      key: "startAt" as const,
      label: "Planned event start",
      value: startAt,
      onChange: setStartAt,
    },
    {
      key: "endAt" as const,
      label: "Planned event end",
      value: endAt,
      onChange: setEndAt,
    },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger render={<Button variant="outline" size="sm" />}>
        <Pencil data-icon="inline-start" /> Edit event plan
      </DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit planned event details</DialogTitle>
          <DialogDescription>
            These details are tentative and are hidden from the public page
            during crowdfunding.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <FieldGroup>
            <Field data-invalid={Boolean(errors.location)}>
              <FieldLabel htmlFor="planned-location">
                Planned location
              </FieldLabel>
              <Input
                id="planned-location"
                value={location}
                onChange={(event) => setLocation(event.target.value)}
                aria-invalid={Boolean(errors.location)}
              />
              {errors.location ? (
                <FieldError>{errors.location}</FieldError>
              ) : null}
            </Field>
            <Field data-invalid={Boolean(errors.description)}>
              <FieldLabel htmlFor="planned-description">
                Short description
              </FieldLabel>
              <Textarea
                id="planned-description"
                value={description}
                onChange={(event) => setDescription(event.target.value)}
                rows={2}
                maxLength={500}
                placeholder="A brief summary for cards and search listings..."
                aria-invalid={Boolean(errors.description)}
              />
              {errors.description ? (
                <FieldError>{errors.description}</FieldError>
              ) : null}
            </Field>
            <Field data-invalid={Boolean(errors.detail)}>
              <FieldLabel htmlFor="planned-detail">
                Tournament Details & Rules (Markdown)
              </FieldLabel>
              <MarkdownEditor
                id="planned-detail"
                value={detail}
                onChange={setDetail}
                rows={6}
                maxLength={10000}
                aria-invalid={Boolean(errors.detail)}
              />
              {errors.detail ? (
                <FieldError>{errors.detail}</FieldError>
              ) : null}
            </Field>
            {fields.map(({ key, label, value, onChange }) => (
              <Field key={key} data-invalid={Boolean(errors[key])}>
                <FieldLabel htmlFor={`planned-${key}`}>{label}</FieldLabel>
                <Input
                  id={`planned-${key}`}
                  type="datetime-local"
                  value={value}
                  onChange={(event) => onChange(event.target.value)}
                  aria-invalid={Boolean(errors[key])}
                />
                {errors[key] ? <FieldError>{errors[key]}</FieldError> : null}
              </Field>
            ))}
          </FieldGroup>
          {apiError ? (
            <Alert variant="destructive">
              <AlertTitle>Could not save</AlertTitle>
              <AlertDescription>{apiError}</AlertDescription>
            </Alert>
          ) : null}
          <DialogFooter>
            <DialogClose
              render={<Button type="button" variant="outline" />}
              disabled={saving}
            >
              Cancel
            </DialogClose>
            <Button type="submit" disabled={saving}>
              {saving ? <Spinner data-icon="inline-start" /> : null}Save plan
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
