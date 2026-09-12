"use client";

import { useState, useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { $api, apiClient } from "@/lib/api/client";
import { extractProblemMessage } from "@/lib/schemas";
import type { components } from "@/lib/api/schema";
import {
  organizerProfileSchema,
  type OrganizerProfileInput,
} from "../schemas";

export type OrganizerProfile =
  components["schemas"]["OrganizerProfileResponse"];

export interface UseOrganizerProfileOptions {
  /**
   * Whether the organizer profile query should execute automatically.
   * Defaults to true.
   */
  enabled?: boolean;
  /**
   * Cache fresh duration in milliseconds.
   * Defaults to 30,000 (30 seconds).
   */
  staleTime?: number;
}

export interface UseOrganizerProfileReturn {
  /** The organizer profile linked to the authenticated user, or null if not yet created. */
  profile: OrganizerProfile | null;
  /** True when either authentication or profile data is resolving. */
  isLoading: boolean;
  /** True when authentication session is loading. */
  isAuthLoading: boolean;
  /** True when the organizer profile query is actively loading. */
  isProfileLoading: boolean;
  /** True when an upsert/save mutation is actively in progress. */
  isSaving: boolean;
  /** Query error if fetching failed (e.g. 404 or network issue). */
  error: unknown;
  /** Error message from the most recent save/upsert attempt, or null. */
  saveError: string | null;
  /** True if the user has an active organizer profile. */
  hasProfile: boolean;
  /** True if the user is authenticated, has an active player account, but lacks an organizer profile. */
  needsOnboarding: boolean;
  /** Upserts the organizer profile details and synchronizes React Query cache. */
  upsertProfile: (
    input: OrganizerProfileInput
  ) => Promise<{ data?: OrganizerProfile; error?: string }>;
  /** Manually refetches the organizer profile. */
  refetch: () => Promise<unknown>;
  /** Clears the save error state. */
  clearSaveError: () => void;
}

/**
 * Hook for managing the current user's organizer profile and onboarding state.
 */
export function useOrganizerProfile(
  options: UseOrganizerProfileOptions = {}
): UseOrganizerProfileReturn {
  const { enabled = true, staleTime = 30_000 } = options;
  const { user, authorizationHeader, isLoading: isAuthLoading } = useAuth();
  const queryClient = useQueryClient();

  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const query = $api.useQuery(
    "get",
    "/api/v1/accounts/me/organizer-profile",
    {
      headers: authorizationHeader
        ? { Authorization: authorizationHeader }
        : {},
    },
    {
      enabled: Boolean(authorizationHeader && enabled),
      retry: false,
      staleTime,
    }
  );

  const {
    data: profileData,
    isLoading: isProfileLoading,
    error,
    refetch,
  } = query;
  const profile = profileData ?? null;
  const hasProfile = Boolean(profile?.id);

  const isLoading =
    isAuthLoading ||
    (Boolean(authorizationHeader && enabled) && isProfileLoading);

  const needsOnboarding = Boolean(
    user && user.status === "ACTIVE" && !isLoading && !hasProfile
  );

  const clearSaveError = useCallback(() => {
    setSaveError(null);
  }, []);

  const upsertProfile = useCallback(
    async (
      input: OrganizerProfileInput
    ): Promise<{ data?: OrganizerProfile; error?: string }> => {
      if (!authorizationHeader) {
        const message = "Sign in before configuring your organizer profile.";
        setSaveError(message);
        return { error: message };
      }

      const validation = organizerProfileSchema.safeParse(input);
      if (!validation.success) {
        const message =
          validation.error.issues[0]?.message ||
          "Invalid organizer profile data.";
        setSaveError(message);
        return { error: message };
      }

      setSaveError(null);
      setIsSaving(true);

      try {
        const { data, error: apiError } = await apiClient.PUT(
          "/api/v1/accounts/me/organizer-profile",
          {
            headers: { Authorization: authorizationHeader },
            body: {
              organizerName: validation.data.organizerName,
              organizerEmail: validation.data.organizerEmail,
            },
          }
        );

        if (apiError || !data) {
          const message = extractProblemMessage(
            apiError,
            "Failed to save organizer profile. Please try again."
          );
          setSaveError(message);
          return { error: message };
        }

        // Instantly sync React Query cache for all organizer profile queries
        queryClient.setQueriesData(
          { queryKey: ["get", "/api/v1/accounts/me/organizer-profile"] },
          data
        );
        await queryClient.invalidateQueries({
          queryKey: ["get", "/api/v1/accounts/me/organizer-profile"],
        });

        return { data, error: undefined };
      } catch (err) {
        const message =
          err instanceof Error
            ? err.message
            : "Connection error. Please check your network and try again.";
        setSaveError(message);
        return { error: message };
      } finally {
        setIsSaving(false);
      }
    },
    [authorizationHeader, queryClient]
  );

  return {
    profile,
    isLoading,
    isAuthLoading,
    isProfileLoading,
    isSaving,
    error,
    saveError,
    hasProfile,
    needsOnboarding,
    upsertProfile,
    refetch,
    clearSaveError,
  };
}
