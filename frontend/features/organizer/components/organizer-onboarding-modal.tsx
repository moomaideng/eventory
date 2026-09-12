"use client";

import { useState, type FormEvent } from "react";
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
  useOrganizerProfile,
  type OrganizerProfile,
} from "../hooks/use-organizer-profile";
import {
  organizerOnboardingSchema,
  type OrganizerOnboardingValues,
} from "../schemas";

type FieldErrors = Partial<Record<keyof OrganizerOnboardingValues, string>>;

interface OrganizerOnboardingFormProps {
  user: UserProfile;
  onSuccess?: (profile: OrganizerProfile) => void;
  onCancel?: () => void;
}

function OrganizerOnboardingForm({
  user,
  onSuccess,
  onCancel,
}: OrganizerOnboardingFormProps) {
  const { logout } = useAuth();
  const { upsertProfile } = useOrganizerProfile();

  const [organizerName, setOrganizerName] = useState(user.displayName || "");
  const [organizerEmail, setOrganizerEmail] = useState(user.email || "");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [apiError, setApiError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setApiError("");

    const validation = organizerOnboardingSchema.safeParse({
      organizerName,
      organizerEmail,
    });

    if (!validation.success) {
      const issues: FieldErrors = {};
      for (const issue of validation.error.issues) {
        const path = issue.path[0] as keyof OrganizerOnboardingValues;
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
      const { data, error } = await upsertProfile({
        organizerName: validation.data.organizerName,
        organizerEmail: validation.data.organizerEmail,
      });

      if (error || !data) {
        setApiError(
          error || "Failed to set up organizer profile. Please try again."
        );
        setIsSubmitting(false);
        return;
      }

      onSuccess?.(data);
    } catch (err) {
      setApiError(
        err instanceof Error
          ? err.message
          : "Unexpected error occurred during organizer onboarding."
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  const isNameValid = organizerName.trim().length > 0;

  return (
    <>
      <AlertDialogHeader>
        <AlertDialogTitle className="text-xl font-bold">
          Set Up Organizer Profile
        </AlertDialogTitle>
        <AlertDialogDescription className="text-sm text-muted-foreground">
          Choose an organization or brand name and contact email to start hosting tournaments.
        </AlertDialogDescription>
      </AlertDialogHeader>

      {/* Connected Account Identity Card */}
      <div className="mt-1 flex items-center justify-between gap-3 rounded-lg border border-border/70 bg-muted/30 p-3">
        <div className="flex min-w-0 items-center gap-3">
          <Avatar size="default" className="border border-border">
            {user.avatarUrl && (
              <AvatarImage src={user.avatarUrl} alt={user.displayName} />
            )}
            <AvatarFallback className="font-semibold text-xs uppercase">
              {user.displayName?.charAt(0) || user.email.charAt(0) || "O"}
            </AvatarFallback>
          </Avatar>
          <div className="flex min-w-0 flex-col leading-tight">
            <span className="text-muted-foreground text-xs font-medium">
              Connected Account
            </span>
            <span className="truncate text-sm font-semibold text-foreground">
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
          {/* Organizer / Organization Name Field */}
          <Field data-invalid={Boolean(fieldErrors.organizerName)}>
            <FieldLabel htmlFor="onboarding-organizer-name">
              Organizer Name
            </FieldLabel>
            <Input
              id="onboarding-organizer-name"
              value={organizerName}
              onChange={(e) => {
                setOrganizerName(e.target.value);
                if (fieldErrors.organizerName) {
                  setFieldErrors((prev) => ({
                    ...prev,
                    organizerName: undefined,
                  }));
                }
                if (apiError) {
                  setApiError("");
                }
              }}
              placeholder="e.g. Chula Esports Club"
              maxLength={150}
              disabled={isSubmitting}
            />
            <FieldDescription>
              Public organization or brand name shown across your tournaments.
            </FieldDescription>
            {fieldErrors.organizerName && (
              <FieldError>{fieldErrors.organizerName}</FieldError>
            )}
          </Field>

          {/* Contact Email Field */}
          <Field data-invalid={Boolean(fieldErrors.organizerEmail)}>
            <FieldLabel htmlFor="onboarding-organizer-email">
              Contact Email
            </FieldLabel>
            <Input
              id="onboarding-organizer-email"
              type="email"
              value={organizerEmail}
              onChange={(e) => {
                setOrganizerEmail(e.target.value);
                if (fieldErrors.organizerEmail) {
                  setFieldErrors((prev) => ({
                    ...prev,
                    organizerEmail: undefined,
                  }));
                }
                if (apiError) {
                  setApiError("");
                }
              }}
              placeholder="contact@chulaesports.com"
              maxLength={255}
              disabled={isSubmitting}
            />
            <FieldDescription>
              Where participants and sponsors can reach you. Defaults to your account email.
            </FieldDescription>
            {fieldErrors.organizerEmail && (
              <FieldError>{fieldErrors.organizerEmail}</FieldError>
            )}
          </Field>
        </FieldGroup>

        {/* API Error Message */}
        {apiError && (
          <div
            role="alert"
            className="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"
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
                Activating Organizer Profile...
              </>
            ) : (
              "Confirm & Start Organizing"
            )}
          </Button>

          {onCancel && (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={onCancel}
              disabled={isSubmitting}
              className="text-muted-foreground hover:text-foreground text-xs"
            >
              Cancel & Return to Tournaments
            </Button>
          )}
        </div>
      </form>
    </>
  );
}

export interface OrganizerOnboardingModalProps {
  /**
   * Explicit controlled open state.
   * If omitted, the modal automatically opens when the authenticated user
   * has an ACTIVE account but lacks an organizer profile.
   */
  open?: boolean;
  /** Callback invoked when modal open state changes. */
  onOpenChange?: (open: boolean) => void;
  /** Callback invoked upon successful profile setup. */
  onSuccess?: (profile: OrganizerProfile) => void;
  /** Callback invoked when the user cancels or dismisses the modal. */
  onCancel?: () => void;
}

export function OrganizerOnboardingModal({
  open,
  onOpenChange,
  onSuccess,
  onCancel,
}: OrganizerOnboardingModalProps = {}) {
  const { user } = useAuth();
  const { needsOnboarding } = useOrganizerProfile();
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
  }

  function handleSuccess(profile: OrganizerProfile) {
    onSuccess?.(profile);
    onOpenChange?.(false);
  }

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
      <AlertDialogContent className="sm:max-w-md">
        <OrganizerOnboardingForm
          key={user.id}
          user={user}
          onSuccess={handleSuccess}
          onCancel={isControlled || onCancel ? handleCancel : undefined}
        />
      </AlertDialogContent>
    </AlertDialog>
  );
}
