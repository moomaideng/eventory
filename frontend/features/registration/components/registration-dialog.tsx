"use client";

import React from "react";
import { $api } from "@/lib/api/client";
import {
  Dialog,
  DialogContent,
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
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { RegistrationQuestion } from "@/features/registration/components/registration-question";
import {
  answerSchema,
  readRegistrationFile,
  type RegistrationSubmission,
} from "@/features/registration/schemas";

export function RegistrationDialog({
  open,
  onOpenChange,
  tournamentId,
  authorizationHeader,
  title,
  submitLabel = "Submit and join",
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tournamentId: string;
  authorizationHeader: string;
  title: string;
  submitLabel?: string;
  onSubmit: (submission: RegistrationSubmission) => Promise<void>;
}) {
  const [busy, setBusy] = React.useState(false);
  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!busy) onOpenChange(next);
      }}
    >
      <DialogContent className="max-h-[85dvh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>
            Complete your individual registration. Your spot is secured after
            submission succeeds.
          </DialogDescription>
        </DialogHeader>
        {open ? (
          <RegistrationDialogForm
            tournamentId={tournamentId}
            authorizationHeader={authorizationHeader}
            onSubmit={onSubmit}
            submitLabel={submitLabel}
            onBusyChange={setBusy}
            onClose={() => onOpenChange(false)}
          />
        ) : null}
      </DialogContent>
    </Dialog>
  );
}

function RegistrationDialogForm({
  tournamentId,
  authorizationHeader,
  onSubmit,
  submitLabel,
  onClose,
  onBusyChange,
}: {
  tournamentId: string;
  authorizationHeader: string;
  submitLabel: string;
  onSubmit: (submission: RegistrationSubmission) => Promise<void>;
  onClose: () => void;
  onBusyChange: (busy: boolean) => void;
}) {
  const [values, setValues] = React.useState<Record<string, string>>({});
  const [files, setFiles] = React.useState<Record<string, File>>({});
  const [consent, setConsent] = React.useState(false);
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [failure, setFailure] = React.useState("");
  const [pending, setPending] = React.useState(false);
  const {
    data: form,
    error,
    isLoading,
    refetch,
  } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/registration-form",
    {
      params: { path: { tournamentId } },
      headers: { Authorization: authorizationHeader },
    },
    { staleTime: 0, retry: false }
  );

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    if (!form || pending) return;
    const result = answerSchema(form.questions ?? []).safeParse({
      values,
      files,
      consent,
    });
    setErrors({});
    setFailure("");
    if (!result.success) {
      setErrors(
        Object.fromEntries(
          result.error.issues.map((issue) => [
            String(issue.path[0]),
            issue.message,
          ])
        )
      );
      return;
    }
    setPending(true);
    onBusyChange(true);
    try {
      const answers = await Promise.all(
        (form.questions ?? []).map(async (q) => ({
          questionId: q.id,
          value: values[q.id] ?? "",
          ...(q.type === "FILE" && files[q.id]
            ? {
                file: {
                  name: files[q.id].name,
                  data: await readRegistrationFile(files[q.id]),
                },
              }
            : {}),
        }))
      );
      await onSubmit({ formVersion: form.version, answers, consent });
      onClose();
    } catch (err) {
      setFailure(
        err instanceof Error
          ? err.message
          : "Could not submit your registration. Please try again."
      );
    } finally {
      setPending(false);
      onBusyChange(false);
    }
  }

  if (isLoading) return <Skeleton className="h-48 w-full" />;
  if (error || !form)
    return (
      <Alert variant="destructive">
        <AlertDescription>
          Could not load registration questions.{" "}
          <Button variant="outline" onClick={() => void refetch()}>
            Try again
          </Button>
        </AlertDescription>
      </Alert>
    );
  return (
    <form onSubmit={submit} className="flex flex-col gap-6" noValidate>
      <fieldset disabled={pending} className="min-w-0">
        <FieldGroup>
          {!(form.questions ?? []).length ? (
            <Alert>
              <AlertDescription>
                The organizer has not added any registration questions for this
                tournament. There is nothing to fill in; only consent is needed.
              </AlertDescription>
            </Alert>
          ) : null}
          {(form.questions ?? []).map((q) => (
            <RegistrationQuestion
              key={q.id}
              question={q}
              value={values[q.id] ?? ""}
              error={errors[q.id]}
              onChange={(value) =>
                setValues((previous) => ({ ...previous, [q.id]: value }))
              }
              onFileChange={(file) =>
                setFiles((previous) => {
                  const next = { ...previous };
                  if (file) next[q.id] = file;
                  else delete next[q.id];
                  return next;
                })
              }
            />
          ))}
          <Field
            orientation="horizontal"
            data-invalid={Boolean(errors.consent)}
          >
            <Checkbox
              id="registration-consent"
              checked={consent}
              onCheckedChange={setConsent}
              aria-invalid={Boolean(errors.consent)}
            />
            <FieldLabel htmlFor="registration-consent">
              {form.consentNotice}
            </FieldLabel>
          </Field>
          {errors.consent ? <FieldError>{errors.consent}</FieldError> : null}
          {errors.files ? <FieldError>{errors.files}</FieldError> : null}
        </FieldGroup>
      </fieldset>
      {failure ? (
        <Alert variant="destructive" role="alert">
          <AlertDescription>{failure}</AlertDescription>
        </Alert>
      ) : null}
      <DialogFooter>
        <Button
          type="button"
          variant="outline"
          disabled={pending}
          onClick={onClose}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={pending}>
          {pending ? <Spinner data-icon="inline-start" /> : null}
          {pending ? "Submitting…" : submitLabel}
        </Button>
      </DialogFooter>
    </form>
  );
}
