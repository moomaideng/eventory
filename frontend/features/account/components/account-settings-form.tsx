"use client";

import { useState, type FormEvent } from "react";
import { CircleCheck } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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

const FIELDS: FieldName[] = ["displayName", "handle", "phone", "avatarUrl"];

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
        phone: flat.phone?.[0],
        avatarUrl: flat.avatarUrl?.[0],
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
        <CardHeader>
          <div className="flex items-center gap-4">
            <Avatar size="lg">
              {baseline.avatarUrl ? (
                <AvatarImage
                  src={baseline.avatarUrl}
                  alt={baseline.displayName}
                />
              ) : null}
              <AvatarFallback>
                {baseline.displayName.slice(0, 2).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <div className="flex flex-col gap-1">
              <CardTitle>Profile</CardTitle>
              <CardDescription>
                How you appear to organizers, sponsors and other players.
              </CardDescription>
            </div>
          </div>
        </CardHeader>

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

            <Field data-invalid={Boolean(fieldErrors.phone)}>
              <FieldLabel htmlFor="account-phone">Phone</FieldLabel>
              <Input
                id="account-phone"
                type="tel"
                autoComplete="tel"
                value={values.phone}
                onChange={(event) => updateField("phone", event.target.value)}
                placeholder="+66 81 234 5678"
                maxLength={20}
                disabled={isSaving}
                aria-invalid={Boolean(fieldErrors.phone)}
              />
              <FieldDescription>Optional.</FieldDescription>
              {fieldErrors.phone ? (
                <FieldError>{fieldErrors.phone}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.avatarUrl)}>
              <FieldLabel htmlFor="account-avatar-url">Avatar URL</FieldLabel>
              <Input
                id="account-avatar-url"
                type="url"
                value={values.avatarUrl}
                onChange={(event) =>
                  updateField("avatarUrl", event.target.value)
                }
                placeholder="https://example.com/me.png"
                disabled={isSaving}
                aria-invalid={Boolean(fieldErrors.avatarUrl)}
              />
              <FieldDescription>
                Optional. A link to an image hosted elsewhere. Leave it empty to
                use your sign-in photo, if you have one.
              </FieldDescription>
              {fieldErrors.avatarUrl ? (
                <FieldError>{fieldErrors.avatarUrl}</FieldError>
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
