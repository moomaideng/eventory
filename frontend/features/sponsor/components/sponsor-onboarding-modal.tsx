"use client";
import { parseSponsorTagDraft } from "@/features/sponsor/tags";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { LogOut } from "lucide-react";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { useAuth, type UserProfile } from "@/context/auth-context";
import {
  useSponsorProfile,
  type SponsorProfile,
} from "../hooks/use-sponsor-profile";
import {
  sponsorOnboardingSchema,
  type SponsorOnboardingValues,
} from "../schemas";
import { uploadSponsorLogo, validateSponsorLogo } from "../logo-upload";

import { SponsorTagsField } from "./sponsor-tags-field";

type FieldErrors = Partial<Record<keyof SponsorOnboardingValues, string>>;

interface SponsorOnboardingFormProps {
  user: UserProfile;
  upsertProfile: (
    input: SponsorOnboardingValues,
  ) => Promise<{ data?: SponsorProfile; error?: string }>;
  onSuccess?: (profile: SponsorProfile) => void;
  onCancel?: () => void;
  showCancelButton?: boolean;
  cancelMessage?: string;
}

function SponsorOnboardingForm({
  user,
  upsertProfile,
  onSuccess,
  onCancel,
  showCancelButton = false,
  cancelMessage = "Cancel",
}: SponsorOnboardingFormProps) {
  const { logout } = useAuth();

  const [sponsorName, setSponsorName] = useState(user.displayName || "");
  const [sponsorEmail, setSponsorEmail] = useState(user.email || "");
  const [websiteUrl, setWebsiteUrl] = useState("");
  const logoUrl = "";
  const [logoFile, setLogoFile] = useState<File | null>(null);
  const [tags, setTags] = useState<string[]>([]);
  const [tagDraft, setTagDraft] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [apiError, setApiError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setApiError("");

    const validation = sponsorOnboardingSchema.safeParse({
      sponsorName,
      sponsorEmail,
      websiteUrl,
      logoUrl,
      tags: [...tags, ...parseSponsorTagDraft(tagDraft)],
    });

    if (!validation.success) {
      const issues: FieldErrors = {};
      for (const issue of validation.error.issues) {
        const path = issue.path[0] as keyof SponsorOnboardingValues;
        if (path && !issues[path]) {
          issues[path] = issue.message;
        }
      }
      setFieldErrors(issues);
      return;
    }

    setFieldErrors({});
    setIsSubmitting(true);

    try {
      const uploadedLogoUrl = logoFile
        ? await uploadSponsorLogo(user.id, logoFile)
        : validation.data.logoUrl;
      const { data, error } = await upsertProfile({
        sponsorName: validation.data.sponsorName,
        sponsorEmail: validation.data.sponsorEmail,
        websiteUrl: validation.data.websiteUrl,
        logoUrl: uploadedLogoUrl,
        tags: validation.data.tags,
      });

      if (error || !data) {
        setApiError(
          error || "Failed to set up sponsor profile. Please try again.",
        );
        setIsSubmitting(false);
        return;
      }

      onSuccess?.(data);
    } catch (err) {
      setApiError(
        err instanceof Error
          ? err.message
          : "Unexpected error occurred during sponsor onboarding.",
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  const isNameValid = sponsorName.trim().length > 0;

  return (
    <>
      <AlertDialogHeader>
        <AlertDialogTitle className="text-xl font-bold">
          Set Up Sponsor Profile
        </AlertDialogTitle>
        <AlertDialogDescription className="text-muted-foreground text-sm">
          Create your sponsor profile to join the directory. Your sponsor name,
          contact email and tags will be visible to organizers.
        </AlertDialogDescription>
      </AlertDialogHeader>

      {/* Connected Account Identity Card */}
      <div className="border-border/70 bg-muted/30 mt-1 flex items-center justify-between gap-3 rounded-lg border p-3">
        <div className="flex min-w-0 items-center gap-3">
          <Avatar size="default" className="border-border border">
            {user.avatarUrl && (
              <AvatarImage src={user.avatarUrl} alt={user.displayName} />
            )}
            <AvatarFallback className="text-xs font-semibold uppercase">
              {user.displayName?.charAt(0) || user.email.charAt(0) || "S"}
            </AvatarFallback>
          </Avatar>
          <div className="flex min-w-0 flex-col leading-tight">
            <span className="text-muted-foreground text-xs font-medium">
              Connected Account
            </span>
            <span className="text-foreground truncate text-sm font-semibold">
              {user.email}
            </span>
          </div>
        </div>

        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={() => logout()}
          disabled={isSubmitting}
          className="text-muted-foreground hover:text-destructive h-8 px-2 text-xs"
          title="Sign out and switch to another account"
        >
          <LogOut className="mr-1 size-3.5" />
          Sign out
        </Button>
      </div>

      <form onSubmit={handleSubmit} className="mt-2 flex flex-col gap-4">
        <FieldGroup>
          {/* Sponsor / Company Name Field */}
          <Field data-invalid={Boolean(fieldErrors.sponsorName)}>
            <FieldLabel htmlFor="onboarding-sponsor-name">
              Sponsor Name
            </FieldLabel>
            <Input
              id="onboarding-sponsor-name"
              value={sponsorName}
              onChange={(e) => {
                setSponsorName(e.target.value);
                if (fieldErrors.sponsorName) {
                  setFieldErrors((prev) => ({
                    ...prev,
                    sponsorName: undefined,
                  }));
                }
                if (apiError) {
                  setApiError("");
                }
              }}
              placeholder="e.g. Acme Corp"
              maxLength={120}
              disabled={isSubmitting}
            />
            {fieldErrors.sponsorName && (
              <FieldError>{fieldErrors.sponsorName}</FieldError>
            )}
          </Field>

          {/* Contact Email Field */}
          <Field data-invalid={Boolean(fieldErrors.sponsorEmail)}>
            <FieldLabel htmlFor="onboarding-sponsor-email">
              Contact Email
            </FieldLabel>
            <Input
              id="onboarding-sponsor-email"
              type="email"
              value={sponsorEmail}
              onChange={(e) => {
                setSponsorEmail(e.target.value);
                if (fieldErrors.sponsorEmail) {
                  setFieldErrors((prev) => ({
                    ...prev,
                    sponsorEmail: undefined,
                  }));
                }
                if (apiError) {
                  setApiError("");
                }
              }}
              placeholder="sponsorships@acme.com"
              maxLength={255}
              disabled={isSubmitting}
            />
            <FieldDescription>
              Where tournament organizers can reach you.
            </FieldDescription>
            {fieldErrors.sponsorEmail && (
              <FieldError>{fieldErrors.sponsorEmail}</FieldError>
            )}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.websiteUrl)}>
            <FieldLabel htmlFor="onboarding-sponsor-website">
              Brand Website (Optional)
            </FieldLabel>
            <Input
              id="onboarding-sponsor-website"
              type="url"
              value={websiteUrl}
              onChange={(e) => {
                setWebsiteUrl(e.target.value);
                if (fieldErrors.websiteUrl) {
                  setFieldErrors((prev) => ({
                    ...prev,
                    websiteUrl: undefined,
                  }));
                }
                if (apiError) setApiError("");
              }}
              placeholder="https://your-brand.com"
              maxLength={2048}
              autoCorrect="off"
              autoCapitalize="none"
              spellCheck={false}
              disabled={isSubmitting}
              aria-invalid={Boolean(fieldErrors.websiteUrl)}
            />
            {fieldErrors.websiteUrl && (
              <FieldError>{fieldErrors.websiteUrl}</FieldError>
            )}
          </Field>

          <Field data-invalid={Boolean(fieldErrors.logoUrl)}>
            <FieldLabel htmlFor="onboarding-sponsor-logo">
              Brand logo (Optional)
            </FieldLabel>
            <Input
              id="onboarding-sponsor-logo"
              type="file"
              accept="image/png,image/jpeg,image/webp"
              onChange={(e) => {
                const file = e.target.files?.[0] ?? null;
                const error = file ? validateSponsorLogo(file) : undefined;
                setLogoFile(error ? null : file);
                setFieldErrors((prev) => ({ ...prev, logoUrl: error }));
                if (apiError) setApiError("");
              }}
              disabled={isSubmitting}
              aria-invalid={Boolean(fieldErrors.logoUrl)}
            />
            <FieldDescription>
              PNG, JPEG, or WebP up to 5 MB. You can replace it later.
            </FieldDescription>
            {fieldErrors.logoUrl && (
              <FieldError>{fieldErrors.logoUrl}</FieldError>
            )}
          </Field>
          <SponsorTagsField
            id="onboarding-sponsor-tags"
            value={tags}
            onChange={setTags}
            draft={tagDraft}
            onDraftChange={setTagDraft}
            error={fieldErrors.tags}
            disabled={isSubmitting}
          />
        </FieldGroup>

        {/* API Error Message */}
        {apiError && (
          <div
            role="alert"
            className="border-destructive/30 bg-destructive/10 text-destructive rounded-lg border p-3 text-sm"
          >
            {apiError}
          </div>
        )}

        {/* Action Buttons */}
        <div className="mt-2 flex flex-col gap-2">
          <Button
            type="submit"
            disabled={isSubmitting || !isNameValid}
            className="w-full font-semibold"
          >
            {isSubmitting ? (
              <>
                <Spinner className="mr-2 size-4" />
                Activating Sponsor Profile...
              </>
            ) : (
              "Confirm & Start Sponsoring"
            )}
          </Button>

          {showCancelButton && onCancel && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={onCancel}
              disabled={isSubmitting}
              className="text-muted-foreground hover:text-foreground text-xs"
            >
              {cancelMessage}
            </Button>
          )}
        </div>
      </form>
    </>
  );
}

export interface SponsorOnboardingModalProps {
  /**
   * Explicit controlled open state.
   * If omitted, the modal automatically opens when the authenticated user
   * has an ACTIVE account but lacks a sponsor profile.
   */
  open?: boolean;
  /** Callback invoked when modal open state changes. */
  onOpenChange?: (open: boolean) => void;
  /** Callback invoked upon successful profile setup. */
  onSuccess?: (profile: SponsorProfile) => void;
  /** Callback invoked when the user cancels or dismisses the modal. */
  onCancel?: () => void;
  /** Optional URL path to redirect to upon modal cancellation (e.g. "/sponsor"). */
  redirectToOnCancel?: string;
  /** Optional to display as the cancel button */
  cancelMessage?: string;
}

export function SponsorOnboardingModal({
  open,
  onOpenChange,
  onSuccess,
  onCancel,
  redirectToOnCancel,
  cancelMessage,
}: SponsorOnboardingModalProps = {}) {
  const router = useRouter();
  const { user } = useAuth();
  const { needsOnboarding, upsertProfile } = useSponsorProfile();
  const [dismissed, setDismissed] = useState(false);

  const isControlled = typeof open === "boolean";
  const isOpen = isControlled ? open : needsOnboarding && !dismissed;

  if (!isOpen || !user) {
    return null;
  }

  function handleCancel() {
    if (!isControlled) {
      setDismissed(true);
    }
    onCancel?.();
    onOpenChange?.(false);
    if (redirectToOnCancel) {
      router.push(redirectToOnCancel);
    }
  }

  function handleSuccess(profile: SponsorProfile) {
    onSuccess?.(profile);
    onOpenChange?.(false);
  }

  const showCancelButton =
    isControlled || Boolean(onCancel) || Boolean(redirectToOnCancel);

  return (
    <AlertDialog
      open={isOpen}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) {
          handleCancel();
        } else {
          onOpenChange?.(true);
        }
      }}
    >
      <AlertDialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-xl">
        <SponsorOnboardingForm
          key={user.id}
          user={user}
          upsertProfile={upsertProfile}
          onSuccess={handleSuccess}
          onCancel={handleCancel}
          showCancelButton={showCancelButton}
          cancelMessage={cancelMessage}
        />
      </AlertDialogContent>
    </AlertDialog>
  );
}
