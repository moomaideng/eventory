"use client";

import { useState, type FormEvent } from "react";
import { CircleCheck } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import {
  accountSettingsSchema,
  type AccountSettingsInput,
} from "@/features/account/schemas";

type FieldName = keyof AccountSettingsInput;
type FieldErrors = Partial<Record<FieldName, string>>;

/** What the manager reports back after trying to save. */
export type SaveResult =
  | { ok: true; values: AccountSettingsInput }
  | { ok: false; message: string; field?: "handle" };

interface AccountSettingsFormProps {
  /** Values currently stored on the account. */
  initial: AccountSettingsInput;
  onSave: (values: AccountSettingsInput) => Promise<SaveResult>;
}

const FIELDS: FieldName[] = ["displayName", "handle"];

// Same cleaning as the onboarding form: drop a leading @, turn spaces into
// underscores, lowercase, and remove anything else the API would not accept.
function sanitizeHandleInput(raw: string): string {
  return raw
    .replace(/^@+/, "")
    .replace(/\s+/g, "_")
    .toLowerCase()
    .replace(/[^a-z0-9_]/g, "");
}

export function AccountSettingsForm({
  initial,
  onSave,
}: AccountSettingsFormProps) {
  const [values, setValues] = useState(initial);
  // Last values known to be stored. Used to tell whether the form was edited.
  const [baseline, setBaseline] = useState(initial);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [apiError, setApiError] = useState("");
  const [showSuccess, setShowSuccess] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const isDirty = FIELDS.some(
    (field) => values[field].trim() !== baseline[field]
  );

  function updateField(field: FieldName, value: string) {
    setValues((prev) => ({ ...prev, [field]: value }));
    setFieldErrors((prev) => ({ ...prev, [field]: undefined }));
    setShowSuccess(false);
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setApiError("");
    setShowSuccess(false);

    const result = accountSettingsSchema.safeParse(values);
    if (!result.success) {
      const flat = result.error.flatten().fieldErrors;
      setFieldErrors({
        displayName: flat.displayName?.[0],
        handle: flat.handle?.[0],
      });
      return;
    }

    setFieldErrors({});
    setIsSaving(true);
    try {
      const outcome = await onSave(result.data);
      if (outcome.ok) {
        // Show what the server actually stored (it may normalise the handle).
        setValues(outcome.values);
        setBaseline(outcome.values);
        setShowSuccess(true);
      } else if (outcome.field === "handle") {
        setFieldErrors({ handle: outcome.message });
      } else {
        setApiError(outcome.message);
      }
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} noValidate>
      <Card>
        <CardContent>
          <FieldGroup>
            <Field data-invalid={Boolean(fieldErrors.displayName)}>
              <FieldLabel htmlFor="account-display-name">
                Display name
              </FieldLabel>
              <Input
                id="account-display-name"
                value={values.displayName}
                onChange={(event) =>
                  updateField("displayName", event.target.value)
                }
                placeholder="e.g. Shadow Ninja"
                maxLength={64}
                disabled={isSaving}
                aria-invalid={Boolean(fieldErrors.displayName)}
              />
              {fieldErrors.displayName ? (
                <FieldError>{fieldErrors.displayName}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.handle)}>
              <FieldLabel htmlFor="account-handle">Handle</FieldLabel>
              <Input
                id="account-handle"
                value={values.handle}
                onChange={(event) =>
                  updateField("handle", sanitizeHandleInput(event.target.value))
                }
                placeholder="your_tag"
                maxLength={32}
                autoCapitalize="none"
                autoCorrect="off"
                spellCheck={false}
                disabled={isSaving}
                aria-invalid={Boolean(fieldErrors.handle)}
              />
              <FieldDescription>
                Your public @handle. 3–32 characters: lowercase letters, numbers
                and underscores.
              </FieldDescription>
              {fieldErrors.handle ? (
                <FieldError>{fieldErrors.handle}</FieldError>
              ) : null}
            </Field>
          </FieldGroup>

          {apiError ? (
            <Alert variant="destructive">
              <AlertTitle>Could not save your changes</AlertTitle>
              <AlertDescription>{apiError}</AlertDescription>
            </Alert>
          ) : null}

          {showSuccess ? (
            <Alert>
              <CircleCheck />
              <AlertTitle>Profile updated</AlertTitle>
              <AlertDescription>
                Your changes now appear across Eventory.
              </AlertDescription>
            </Alert>
          ) : null}
        </CardContent>

        <CardFooter className="justify-end">
          <Button type="submit" disabled={isSaving || !isDirty}>
            {isSaving ? <Spinner data-icon="inline-start" /> : null}
            {isSaving ? "Saving…" : "Save changes"}
          </Button>
        </CardFooter>
      </Card>
    </form>
  );
}
