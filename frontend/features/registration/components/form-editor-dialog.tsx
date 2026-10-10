"use client";

import { useState } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
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
import { Textarea } from "@/components/ui/textarea";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Spinner } from "@/components/ui/spinner";
import { QuestionEditor } from "@/features/registration/components/question-editor";
import {
  registrationConfigSchema,
  newRegistrationQuestion,
  type RegistrationConfig,
} from "@/features/registration/schemas";

export function FormEditorDialog({
  open,
  onOpenChange,
  initial,
  onSave,
  draft = false,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initial: RegistrationConfig;
  onSave: (config: RegistrationConfig) => Promise<void>;
  draft?: boolean;
}) {
  const [busy, setBusy] = useState(false);
  return (
    <Dialog
      open={open}
      onOpenChange={(value) => {
        if (!busy) onOpenChange(value);
      }}
    >
      <DialogContent className="max-h-[85dvh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Registration form</DialogTitle>
          <DialogDescription>
            {draft
              ? "Choose what each participant should fill in. This form will be saved when you publish the tournament."
              : "Changes apply to new applicants only. Existing participants keep their original answers and do not need to submit again."}
          </DialogDescription>
        </DialogHeader>
        {open ? (
          <FormEditor
            initial={initial}
            onSave={onSave}
            onClose={() => onOpenChange(false)}
            onBusyChange={setBusy}
            draft={draft}
          />
        ) : null}
      </DialogContent>
    </Dialog>
  );
}

function FormEditor({
  initial,
  onSave,
  onClose,
  onBusyChange,
  draft,
}: {
  initial: RegistrationConfig;
  onSave: (config: RegistrationConfig) => Promise<void>;
  onClose: () => void;
  onBusyChange: (busy: boolean) => void;
  draft: boolean;
}) {
  const [config, setConfig] = useState(initial);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [failure, setFailure] = useState("");
  const [pending, setPending] = useState(false);
  async function save() {
    const result = registrationConfigSchema.safeParse(config);
    setErrors({});
    setFailure("");
    if (!result.success) {
      setErrors(
        Object.fromEntries(
          result.error.issues.map((i) => [i.path.join("."), i.message])
        )
      );
      return;
    }
    setPending(true);
    onBusyChange(true);
    try {
      await onSave(result.data);
      onClose();
    } catch (e) {
      setFailure(
        e instanceof Error ? e.message : "Could not save the form. Try again."
      );
    } finally {
      setPending(false);
      onBusyChange(false);
    }
  }
  // No nested HTML form: this editor can open inside the host-tournament form.
  return (
    <div
      className="flex flex-col gap-6"
      onKeyDown={(event) => {
        if (event.key === "Enter" && event.target instanceof HTMLInputElement)
          event.preventDefault();
      }}
    >
      <fieldset disabled={pending} className="min-w-0">
        <FieldGroup>
          {!config.questions.length ? (
            <Alert>
              <AlertDescription>
                No questions yet. Participants will only be asked for consent.
              </AlertDescription>
            </Alert>
          ) : null}
          {config.questions.map((q, index) => (
            <QuestionEditor
              key={q.id}
              question={q}
              index={index}
              errors={Object.fromEntries(
                Object.entries(errors)
                  .filter(([key]) => key.startsWith(`questions.${index}.`))
                  .map(([key, value]) => [key.split(".")[2], value])
              )}
              onChange={(next) =>
                setConfig((prev) => ({
                  ...prev,
                  questions: prev.questions.map((old) =>
                    old.id === q.id ? next : old
                  ),
                }))
              }
              onRemove={() => {
                setErrors({});
                setConfig((prev) => ({
                  ...prev,
                  questions: prev.questions.filter((old) => old.id !== q.id),
                }));
              }}
            />
          ))}
          <Button
            type="button"
            variant="outline"
            disabled={config.questions.length >= 30}
            onClick={() =>
              setConfig((prev) => ({
                ...prev,
                questions: [...prev.questions, newRegistrationQuestion()],
              }))
            }
          >
            <Plus data-icon="inline-start" />
            Add question
          </Button>
          {errors.questions ? (
            <FieldError>{errors.questions}</FieldError>
          ) : null}
          <Field data-invalid={Boolean(errors.consentNotice)}>
            <FieldLabel htmlFor="form-consent-notice">
              Participant consent notice
            </FieldLabel>
            <Textarea
              id="form-consent-notice"
              maxLength={2000}
              value={config.consentNotice}
              aria-invalid={Boolean(errors.consentNotice)}
              onChange={(e) =>
                setConfig((prev) => ({
                  ...prev,
                  consentNotice: e.target.value,
                }))
              }
            />
            {errors.consentNotice ? (
              <FieldError>{errors.consentNotice}</FieldError>
            ) : null}
          </Field>
        </FieldGroup>
      </fieldset>
      {failure ? (
        <Alert variant="destructive">
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
        <Button type="button" disabled={pending} onClick={() => void save()}>
          {pending ? <Spinner data-icon="inline-start" /> : null}
          {pending ? "Saving…" : draft ? "Use this form" : "Save form"}
        </Button>
      </DialogFooter>
    </div>
  );
}
