"use client";

import { useState, type FormEvent } from "react";
import { LogOut } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
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
import { apiClient } from "@/lib/api/client";
import { extractProblemMessage } from "@/lib/schemas";
import {
  onboardingFormSchema,
  type OnboardingFormValues,
} from "../schemas";

type FieldErrors = Partial<Record<keyof OnboardingFormValues, string>>;

/**
 * Derives the smartest initial handle:
 * 1. Uses backend handle if it's not a generic fallback (like user_xxx or player_xxx).
 * 2. Otherwise sanitizes email username prefix (e.g. somchai.pro -> somchai_pro).
 * 3. Falls back to backend handle.
 */
function getSmartInitialHandle(user: UserProfile): string {
  if (
    user.handle &&
    !user.handle.startsWith("user_") &&
    !user.handle.startsWith("player_")
  ) {
    return user.handle;
  }

  if (user.email) {
    const emailPrefix = user.email
      .split("@")[0]
      .toLowerCase()
      .replace(/[^a-z0-9_]/g, "_")
      .replace(/^_+|_+$/g, "")
      .slice(0, 32);

    if (emailPrefix.length >= 3) {
      return emailPrefix;
    }
  }

  return user.handle || "";
}

function OnboardingForm({ user }: { user: UserProfile }) {
  const { authorizationHeader, updateUserProfile, logout } = useAuth();
  const queryClient = useQueryClient();

  const [displayName, setDisplayName] = useState(user.displayName || "");
  const [handle, setHandle] = useState(() => getSmartInitialHandle(user));
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [apiError, setApiError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Clean and sanitize handle input in real-time
  function handleHandleChange(raw: string) {
    const sanitized = raw
      .replace(/^@+/, "") // Remove leading @ if typed or pasted
      .replace(/\s+/g, "_") // Convert spaces to underscores
      .toLowerCase()
      .replace(/[^a-z0-9_]/g, ""); // Keep only lowercase alphanumeric + underscore

    setHandle(sanitized);
    if (fieldErrors.handle) {
      setFieldErrors((prev) => ({ ...prev, handle: undefined }));
    }
    if (apiError) {
      setApiError("");
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setApiError("");

    const validation = onboardingFormSchema.safeParse({
      displayName,
      handle,
    });

    if (!validation.success) {
      const issues: FieldErrors = {};
      for (const issue of validation.error.issues) {
        const path = issue.path[0] as keyof OnboardingFormValues;
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
      if (!authorizationHeader) {
        setApiError("Authentication session expired. Please sign in again.");
        setIsSubmitting(false);
        return;
      }

      const { data, error } = await apiClient.PATCH("/api/v1/accounts/me", {
        headers: {
          Authorization: authorizationHeader,
        },
        body: {
          displayName: validation.data.displayName,
          handle: validation.data.handle,
        },
      });

      if (error) {
        const message = extractProblemMessage(
          error,
          "Failed to activate player profile. Please try another handle."
        );
        setApiError(message);
        setIsSubmitting(false);
        return;
      }

      if (data) {
        // Instantly update React Query cache and Auth context
        updateUserProfile(
          data.displayName,
          data.handle,
          data.avatarUrl ?? undefined,
          data.phone ?? undefined,
          "ACTIVE"
        );
        queryClient.setQueryData(["get", "/api/v1/accounts/me"], data);
        await queryClient.invalidateQueries({
          queryKey: ["get", "/api/v1/accounts/me"],
        });
      }
    } catch (err) {
      setApiError(
        err instanceof Error
          ? err.message
          : "Unexpected error occurred during onboarding."
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  const isHandleLengthValid = handle.length >= 3 && handle.length <= 32;

  return (
    <>
      <AlertDialogHeader>
        <AlertDialogTitle className="text-xl font-bold">
          Set Up Your Profile
        </AlertDialogTitle>
        <AlertDialogDescription className="text-sm text-muted-foreground">
          Choose your display name and handle to get started.
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
              {user.displayName?.charAt(0) || user.email.charAt(0) || "P"}
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
          {/* Display Name Field (Pre-filled from Google) */}
          <Field data-invalid={Boolean(fieldErrors.displayName)}>
            <FieldLabel htmlFor="onboarding-display-name">
              Display Name
            </FieldLabel>
            <Input
              id="onboarding-display-name"
              value={displayName}
              onChange={(e) => {
                setDisplayName(e.target.value);
                if (fieldErrors.displayName) {
                  setFieldErrors((prev) => ({
                    ...prev,
                    displayName: undefined,
                  }));
                }
              }}
              placeholder="e.g. Shadow Ninja"
              maxLength={64}
              disabled={isSubmitting}
            />
            {fieldErrors.displayName && (
              <FieldError>{fieldErrors.displayName}</FieldError>
            )}
          </Field>

          {/* Player Handle Field */}
          <Field data-invalid={Boolean(fieldErrors.handle)}>
            <FieldLabel htmlFor="onboarding-handle">
              Handle
            </FieldLabel>

            <div className="relative flex items-center">
              <span className="text-muted-foreground pointer-events-none absolute left-3 select-none text-sm font-semibold">
                @
              </span>
              <Input
                id="onboarding-handle"
                className="pl-7 lowercase font-mono tracking-tight"
                value={handle}
                onChange={(e) => handleHandleChange(e.target.value)}
                placeholder="your_tag"
                maxLength={32}
                disabled={isSubmitting}
              />
            </div>
            <FieldDescription>
              3–32 characters (lowercase letters, numbers, and underscores).
            </FieldDescription>
            {fieldErrors.handle && (
              <FieldError>{fieldErrors.handle}</FieldError>
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
            disabled={isSubmitting || !displayName.trim() || !isHandleLengthValid}
            className="w-full font-semibold"
          >
            {isSubmitting ? (
              <>
                <Spinner className="mr-2 size-4" />
                Activating Profile...
              </>
            ) : (
              "Confirm & Start Playing"
            )}
          </Button>

          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => logout()}
            disabled={isSubmitting}
            className="text-muted-foreground hover:text-foreground text-xs"
          >
            Cancel & Browse as Guest
          </Button>
        </div>
      </form>
    </>
  );
}

export function OnboardingModal() {
  const { user } = useAuth();
  const isOpen = Boolean(user && user.status === "ONBOARDING");

  if (!isOpen || !user) {
    return null;
  }

  return (
    <AlertDialog open={isOpen}>
      <AlertDialogContent className="sm:max-w-md">
        <OnboardingForm key={user.id} user={user} />
      </AlertDialogContent>
    </AlertDialog>
  );
}
