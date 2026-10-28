"use client";

import {
  useState,
  type FormEvent,
  type MouseEvent as ReactMouseEvent,
} from "react";
import { Building2 } from "lucide-react";
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
import { sponsorProfileSchema, type SponsorProfileInput } from "../schemas";

type FieldErrors = Partial<
  Record<
    | "sponsorName"
    | "sponsorEmail"
    | "category"
    | "publicContactEmail"
    | "directoryVisible",
    string
  >
>;

interface SponsorProfileFormProps {
  initialName: string;
  initialEmail: string;
  initialCategory: string;
  initialPublicContactEmail: string;
  initialDirectoryVisible: boolean;
  onCancel: () => void;
  onConfirmSave: (data: SponsorProfileInput) => Promise<boolean>;
  isSaving: boolean;
  saveError: string;
}

export function SponsorProfileForm({
  initialName,
  initialEmail,
  initialCategory,
  initialPublicContactEmail,
  initialDirectoryVisible,
  onCancel,
  onConfirmSave,
  isSaving,
  saveError,
}: SponsorProfileFormProps) {
  const [sponsorName, setSponsorName] = useState(initialName);
  const [sponsorEmail, setSponsorEmail] = useState(initialEmail);
  const [category, setCategory] = useState(initialCategory);
  const [publicContactEmail, setPublicContactEmail] = useState(
    initialPublicContactEmail
  );
  const [directoryVisible, setDirectoryVisible] = useState(
    initialDirectoryVisible
  );
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [dialogOpen, setDialogOpen] = useState(false);
  const [pendingData, setPendingData] = useState<SponsorProfileInput | null>(
    null
  );

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const result = sponsorProfileSchema.safeParse({
      sponsorName,
      sponsorEmail,
      category,
      publicContactEmail,
      directoryVisible,
    });

    if (!result.success) {
      const flat = result.error.flatten().fieldErrors;
      setFieldErrors({
        sponsorName: flat.sponsorName?.[0],
        sponsorEmail: flat.sponsorEmail?.[0],
        category: flat.category?.[0],
        publicContactEmail: flat.publicContactEmail?.[0],
        directoryVisible: flat.directoryVisible?.[0],
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
        <CardTitle>Sponsor profile</CardTitle>
        <CardDescription>
          This information is shown to organizers when you sponsor a tournament.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit} noValidate>
        <CardContent>
          <FieldGroup>
            <Field data-invalid={Boolean(fieldErrors.sponsorName)}>
              <FieldLabel htmlFor="sponsor-name">Sponsor name</FieldLabel>
              <Input
                id="sponsor-name"
                value={sponsorName}
                onChange={(event) => {
                  setSponsorName(event.target.value);
                  setFieldErrors((prev) => ({
                    ...prev,
                    sponsorName: undefined,
                  }));
                }}
                placeholder="Acme Corp"
                maxLength={120}
                aria-invalid={Boolean(fieldErrors.sponsorName)}
              />
              {fieldErrors.sponsorName ? (
                <FieldError>{fieldErrors.sponsorName}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.sponsorEmail)}>
              <FieldLabel htmlFor="sponsor-email">Contact email</FieldLabel>
              <Input
                id="sponsor-email"
                type="email"
                value={sponsorEmail}
                onChange={(event) => {
                  setSponsorEmail(event.target.value);
                  setFieldErrors((prev) => ({
                    ...prev,
                    sponsorEmail: undefined,
                  }));
                }}
                placeholder="sponsorships@acme.com"
                aria-invalid={Boolean(fieldErrors.sponsorEmail)}
              />
              <FieldDescription>Optional.</FieldDescription>
              {fieldErrors.sponsorEmail ? (
                <FieldError>{fieldErrors.sponsorEmail}</FieldError>
              ) : null}
            </Field>
            <div className="border-border space-y-4 rounded-lg border p-4">
              <div>
                <p className="font-medium">Sponsor directory</p>
                <p className="text-muted-foreground text-sm">
                  Organizers can find you only when you choose to appear here.
                  New profiles start hidden.
                </p>
              </div>
              <Field data-invalid={Boolean(fieldErrors.category)}>
                <FieldLabel htmlFor="sponsor-category">
                  Business category
                </FieldLabel>
                <Input
                  id="sponsor-category"
                  value={category}
                  maxLength={80}
                  placeholder="e.g. Technology"
                  aria-invalid={Boolean(fieldErrors.category)}
                  onChange={(event) => {
                    setCategory(event.target.value);
                    setFieldErrors((prev) => ({
                      ...prev,
                      category: undefined,
                    }));
                  }}
                />
                {fieldErrors.category ? (
                  <FieldError>{fieldErrors.category}</FieldError>
                ) : null}
              </Field>
              <Field data-invalid={Boolean(fieldErrors.publicContactEmail)}>
                <FieldLabel htmlFor="public-contact-email">
                  Public contact email
                </FieldLabel>
                <Input
                  id="public-contact-email"
                  type="email"
                  value={publicContactEmail}
                  placeholder="partners@example.com"
                  aria-invalid={Boolean(fieldErrors.publicContactEmail)}
                  onChange={(event) => {
                    setPublicContactEmail(event.target.value);
                    setFieldErrors((prev) => ({
                      ...prev,
                      publicContactEmail: undefined,
                    }));
                  }}
                />
                <FieldDescription>
                  This address appears in the directory. Your profile contact
                  email above stays separate.
                </FieldDescription>
                {fieldErrors.publicContactEmail ? (
                  <FieldError>{fieldErrors.publicContactEmail}</FieldError>
                ) : null}
              </Field>
              <label
                className="flex items-start gap-3 text-sm"
                htmlFor="directory-visible"
              >
                <input
                  id="directory-visible"
                  type="checkbox"
                  checked={directoryVisible}
                  className="border-input accent-primary mt-0.5 size-4 rounded border"
                  onChange={(event) => {
                    setDirectoryVisible(event.target.checked);
                    setFieldErrors((prev) => ({
                      ...prev,
                      directoryVisible: undefined,
                    }));
                  }}
                />
                <span>
                  Show my sponsor name, category and public contact email to
                  organizers
                </span>
              </label>
              {fieldErrors.directoryVisible ? (
                <FieldError>{fieldErrors.directoryVisible}</FieldError>
              ) : null}
            </div>
          </FieldGroup>
        </CardContent>
        <CardFooter className="justify-end gap-3">
          <Button type="button" variant="ghost" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit">
            <Building2 data-icon="inline-start" />
            Save profile
          </Button>
        </CardFooter>
      </form>

      <AlertDialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Save changes?</AlertDialogTitle>
            <AlertDialogDescription>
              {pendingData?.directoryVisible
                ? "Your sponsor name, category and public contact email will be available to organizers in the directory."
                : "Your sponsor profile will be saved. Your directory listing will remain hidden."}
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
