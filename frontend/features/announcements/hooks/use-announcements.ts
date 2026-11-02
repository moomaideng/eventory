"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import type { components } from "@/lib/api/schema";
import { extractProblemMessage } from "@/lib/schemas";
import type { AnnouncementInput } from "@/features/announcements/schemas";

export type Announcement = components["schemas"]["AnnouncementResponse"];

const ANNOUNCEMENTS_PATH = "/api/v1/tournaments/{id}/announcements";

/**
 * Reads a tournament's announcements, newest first. The list is public, so no
 * sign-in header is sent. Pass enabled=false to skip the request, for example
 * while the tournament is not published and the API would answer 404.
 */
export function useAnnouncements(tournamentId: string, enabled = true) {
  const query = $api.useQuery(
    "get",
    ANNOUNCEMENTS_PATH,
    { params: { path: { id: tournamentId } } },
    { enabled }
  );

  return {
    // The API types the list as nullable, so fall back to an empty one.
    announcements: query.data?.items ?? [],
    isLoading: query.isLoading,
    isError: query.isError,
    refetch: query.refetch,
  };
}

/**
 * Posts an announcement. The server owns who may post, so a refusal comes back
 * as a message for the form to show. On success the list is refreshed so the
 * new announcement appears at the top.
 */
export function usePostAnnouncement(tournamentId: string) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();
  const mutation = $api.useMutation("post", ANNOUNCEMENTS_PATH);

  function post(
    input: AnnouncementInput,
    options: { onSuccess?: () => void } = {}
  ) {
    mutation.mutate(
      {
        params: { path: { id: tournamentId } },
        headers: { Authorization: authorizationHeader ?? "" },
        body: input,
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries({
            queryKey: ["get", ANNOUNCEMENTS_PATH],
          });
          options.onSuccess?.();
        },
      }
    );
  }

  return {
    post,
    isPending: mutation.isPending,
    errorMessage: mutation.isError
      ? extractProblemMessage(
          mutation.error,
          "Could not post the announcement. Please try again."
        )
      : null,
    reset: mutation.reset,
  };
}
