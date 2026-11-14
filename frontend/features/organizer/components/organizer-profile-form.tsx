"use client";

import {
  useState,
  type FormEvent,
  type MouseEvent as ReactMouseEvent,
} from "react";
import { Trophy } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
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
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import {
  organizerCreateSchema,
  organizerProfileSchema,
  type OrganizerProfileInput,
} from "../schemas";
import { PrivacyConsentField } from "@/components/privacy-consent-field";

type FieldErrors = Partial<
  Record<
    "organizerName" | "organizerEmail" | "phone" | "privacyConsentAccepted",
    string
  >
>;

interface OrganizerProfileFormProps {
  initialName: string;
  initialEmail: string;
  initialPhone: string;
  isCreating?: boolean;
  onCancel: () => void;
  onConfirmSave: (data: OrganizerProfileInput) => Promise<boolean>;
  isSaving: boolean;
  saveError: string;
}

export function OrganizerProfileForm({
  initialName,
  initialEmail,
  initialPhone,
  isCreating = false,
  onCancel,
  onConfirmSave,
  isSaving,
  saveError,
}: OrganizerProfileFormProps) {
  const [organizerName, setOrganizerName] = useState(initialName);
  const [organizerEmail, setOrganizerEmail] = useState(initialEmail);
  const [phone, setPhone] = useState(initialPhone);
  const [privacyConsentAccepted, setPrivacyConsentAccepted] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [dialogOpen, setDialogOpen] = useState(false);
  const [pendingData, setPendingData] = useState<OrganizerProfileInput | null>(
    null
  );

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const schema = isCreating ? organizerCreateSchema : organizerProfileSchema;
    const result = schema.safeParse({
      organizerName,
      organizerEmail,
      phone,
      privacyConsentAccepted: isCreating ? privacyConsentAccepted : undefined,
    });

    if (!result.success) {
      const flat = result.error.flatten().fieldErrors;
      setFieldErrors({
        organizerName: flat.organizerName?.[0],
        organizerEmail: flat.organizerEmail?.[0],
        phone: flat.phone?.[0],
        privacyConsentAccepted: flat.privacyConsentAccepted?.[0],
      });
      return;
    }

    setFieldErrors({});
    setPendingData(result.data);
    setDialogOpen(true);
  }

  async function handleConfirm(event: ReactMouseEvent) {
    // AlertDialogAction closes the dialog on click by default. Save is
    // async, so we take over closing ourselves: stay open on failure so
    // the error message (and retry) stay visible.
    event.preventDefault();
    if (!pendingData) return;
    const ok = await onConfirmSave(pendingData);
    if (ok) {
      setDialogOpen(false);
      setPendingData(null);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Organizer profile</CardTitle>
        <CardDescription>
          This information is shown to sponsors and participants across your
          tournaments.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit} noValidate>
        <CardContent>
          <FieldGroup>
            <Field data-invalid={Boolean(fieldErrors.organizerName)}>
              <FieldLabel htmlFor="organizer-name">Organizer name</FieldLabel>
              <Input
                id="organizer-name"
                value={organizerName}
                onChange={(event) => {
                  setOrganizerName(event.target.value);
                  setFieldErrors((prev) => ({
                    ...prev,
                    organizerName: undefined,
                  }));
                }}
                placeholder="Chula Esports Club"
                maxLength={150}
                aria-invalid={Boolean(fieldErrors.organizerName)}
              />
              {fieldErrors.organizerName ? (
                <FieldError>{fieldErrors.organizerName}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.organizerEmail)}>
              <FieldLabel htmlFor="organizer-email">Contact email</FieldLabel>
              <Input
                id="organizer-email"
                type="email"
                value={organizerEmail}
                onChange={(event) => {
                  setOrganizerEmail(event.target.value);
                  setFieldErrors((prev) => ({
                    ...prev,
                    organizerEmail: undefined,
                  }));
                }}
                placeholder="contact@chulaesports.com"
                aria-invalid={Boolean(fieldErrors.organizerEmail)}
              />
              <FieldDescription>Optional.</FieldDescription>
              {fieldErrors.organizerEmail ? (
                <FieldError>{fieldErrors.organizerEmail}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.phone)}>
              <FieldLabel htmlFor="organizer-phone">Phone</FieldLabel>
              <Input
                id="organizer-phone"
                type="tel"
                value={phone}
                onChange={(event) => {
                  setPhone(event.target.value);
                  setFieldErrors((prev) => ({ ...prev, phone: undefined }));
                }}
                placeholder="+66 81 234 5678"
                maxLength={20}
                aria-invalid={Boolean(fieldErrors.phone)}
              />
              <FieldDescription>Optional.</FieldDescription>
              {fieldErrors.phone ? (
                <FieldError>{fieldErrors.phone}</FieldError>
              ) : null}
            </Field>

            {isCreating ? (
              <PrivacyConsentField
                id="settings-organizer-privacy-consent"
                checked={privacyConsentAccepted}
                onCheckedChange={(checked) => {
                  setPrivacyConsentAccepted(checked);
                  if (checked && fieldErrors.privacyConsentAccepted) {
                    setFieldErrors((prev) => ({
                      ...prev,
                      privacyConsentAccepted: undefined,
                    }));
                  }
                }}
                description="Your organizer name and contact details will be saved to our database and displayed to participants and sponsors on tournaments you host."
                error={fieldErrors.privacyConsentAccepted}
              />
            ) : null}
          </FieldGroup>
        </CardContent>
        <CardFooter className="justify-end gap-3">
          <Button type="button" variant="ghost" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit">
            <Trophy data-icon="inline-start" />
            Save profile
          </Button>
        </CardFooter>
      </form>

      <AlertDialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Save changes?</AlertDialogTitle>
            <AlertDialogDescription>
              Sponsors and participants will see this organizer name and email
              the next time they view your tournaments.
            </AlertDialogDescription>
          </AlertDialogHeader>
          {saveError ? (
            <p className="text-destructive text-sm">{saveError}</p>
          ) : null}
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isSaving}>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirm} disabled={isSaving}>
              {isSaving ? <Spinner data-icon="inline-start" /> : null}
              {isSaving ? "Saving…" : "Confirm save"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
}
