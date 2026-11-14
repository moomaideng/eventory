"use client";
import { parseSponsorTagDraft } from "@/features/sponsor/tags";

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
import { validateSponsorLogo } from "../logo-upload";

type FieldErrors = Partial<
  Record<
    | "sponsorName"
    | "sponsorEmail"
    | "phone"
    | "websiteUrl"
    | "logoUrl"
    | "tags",
    string
  >
>;

interface SponsorProfileFormProps {
  initialName: string;
  initialEmail: string;
  initialPhone: string;
  initialWebsiteUrl: string;
  initialLogoUrl: string;
  initialTags: string[];
  onCancel: () => void;
  onConfirmSave: (
    data: SponsorProfileInput,
    logoFile: File | null,
  ) => Promise<boolean>;
  isSaving: boolean;
  saveError: string;
}

export function SponsorProfileForm({
  initialName,
  initialEmail,
  initialPhone,
  initialWebsiteUrl,
  initialLogoUrl,
  initialTags,
  onCancel,
  onConfirmSave,
  isSaving,
  saveError,
}: SponsorProfileFormProps) {
  const [sponsorName, setSponsorName] = useState(initialName);
  const [sponsorEmail, setSponsorEmail] = useState(initialEmail);
  const [phone, setPhone] = useState(initialPhone);
  const [websiteUrl, setWebsiteUrl] = useState(initialWebsiteUrl);
  const logoUrl = initialLogoUrl;
  const [logoFile, setLogoFile] = useState<File | null>(null);
  const [tags, setTags] = useState(initialTags);
  const [tagDraft, setTagDraft] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [dialogOpen, setDialogOpen] = useState(false);
  const [pendingData, setPendingData] = useState<SponsorProfileInput | null>(
    null,
  );

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const result = sponsorProfileSchema.safeParse({
      sponsorName,
      sponsorEmail,
      phone,
      websiteUrl,
      logoUrl,
      tags: [...tags, ...parseSponsorTagDraft(tagDraft)],
    });

    if (!result.success) {
      const flat = result.error.flatten().fieldErrors;
      setFieldErrors({
        sponsorName: flat.sponsorName?.[0],
        sponsorEmail: flat.sponsorEmail?.[0],
        phone: flat.phone?.[0],
        websiteUrl: flat.websiteUrl?.[0],
        logoUrl: flat.logoUrl?.[0],
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
    const ok = await onConfirmSave(pendingData, logoFile);
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

            <Field data-invalid={Boolean(fieldErrors.phone)}>
              <FieldLabel htmlFor="sponsor-phone">Phone</FieldLabel>
              <Input
                id="sponsor-phone"
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

            <Field data-invalid={Boolean(fieldErrors.websiteUrl)}>
              <FieldLabel htmlFor="sponsor-website-url">
                Brand website
              </FieldLabel>
              <Input
                id="sponsor-website-url"
                type="url"
                value={websiteUrl}
                onChange={(event) => {
                  setWebsiteUrl(event.target.value);
                  setFieldErrors((prev) => ({
                    ...prev,
                    websiteUrl: undefined,
                  }));
                }}
                placeholder="https://your-brand.com"
                maxLength={2048}
                autoCorrect="off"
                autoCapitalize="none"
                spellCheck={false}
                aria-invalid={Boolean(fieldErrors.websiteUrl)}
                aria-describedby={
                  fieldErrors.websiteUrl
                    ? "sponsor-website-error"
                    : "sponsor-website-hint"
                }
              />
              <FieldDescription id="sponsor-website-hint">
                Required before sponsoring a tournament. Your logo links here.
              </FieldDescription>
              {fieldErrors.websiteUrl ? (
                <FieldError id="sponsor-website-error">
                  {fieldErrors.websiteUrl}
                </FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.logoUrl)}>
              <FieldLabel htmlFor="sponsor-logo">Brand logo</FieldLabel>
              <Input
                id="sponsor-logo"
                type="file"
                accept="image/png,image/jpeg,image/webp"
                onChange={(event) => {
                  const file = event.target.files?.[0] ?? null;
                  const error = file ? validateSponsorLogo(file) : undefined;
                  setLogoFile(error ? null : file);
                  setFieldErrors((prev) => ({ ...prev, logoUrl: error }));
                }}
                aria-invalid={Boolean(fieldErrors.logoUrl)}
                aria-describedby={
                  fieldErrors.logoUrl
                    ? "sponsor-logo-error"
                    : "sponsor-logo-hint"
                }
              />
              <FieldDescription id="sponsor-logo-hint">
                PNG, JPEG, or WebP up to 5 MB. A new upload replaces your
                current logo.
              </FieldDescription>
              {fieldErrors.logoUrl ? (
                <FieldError id="sponsor-logo-error">
                  {fieldErrors.logoUrl}
                </FieldError>
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
