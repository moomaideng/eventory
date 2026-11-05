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
import { SponsorTagsField } from "./sponsor-tags-field";
import { sponsorProfileSchema, type SponsorProfileInput } from "../schemas";

type FieldErrors = Partial<
  Record<"sponsorName" | "sponsorEmail" | "tags", string>
>;

interface SponsorProfileFormProps {
  initialName: string;
  initialEmail: string;
  initialTags: string[];
  onCancel: () => void;
  onConfirmSave: (data: SponsorProfileInput) => Promise<boolean>;
  isSaving: boolean;
  saveError: string;
}

export function SponsorProfileForm({
  initialName,
  initialEmail,
  initialTags,
  onCancel,
  onConfirmSave,
  isSaving,
  saveError,
}: SponsorProfileFormProps) {
  const [sponsorName, setSponsorName] = useState(initialName);
  const [sponsorEmail, setSponsorEmail] = useState(initialEmail);
  const [tags, setTags] = useState(initialTags);
  const [tagDraft, setTagDraft] = useState("");
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
      tags: [...tags, tagDraft],
    });

    if (!result.success) {
      const flat = result.error.flatten().fieldErrors;
      setFieldErrors({
        sponsorName: flat.sponsorName?.[0],
        sponsorEmail: flat.sponsorEmail?.[0],
        tags: flat.tags?.[0],
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
          Your sponsor name, contact email and tags are visible to organizers in
          the sponsor directory.
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
              <FieldDescription>
                Organizers can use this email to contact you from the sponsor
                directory.
              </FieldDescription>
              {fieldErrors.sponsorEmail ? (
                <FieldError>{fieldErrors.sponsorEmail}</FieldError>
              ) : null}
            </Field>
            <SponsorTagsField
              value={tags}
              onChange={setTags}
              draft={tagDraft}
              onDraftChange={setTagDraft}
              error={fieldErrors.tags}
            />
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
              Your sponsor name, tags and contact email will be available to
              organizers in the directory.
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
