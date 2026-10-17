"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { extractProblemMessage } from "@/lib/schemas";
import type { components } from "@/lib/api/schema";

type ReviewDecision = components["schemas"]["TeamReviewInputBody"]["decision"];

/**
 * Accepts or rejects one application (a solo player is a one-person team, so
 * one endpoint covers both). The server owns every rule (stage, capacity,
 * allowed transitions), so a refusal comes back as a message for the dialog
 * to show rather than being second-guessed here.
 */
export function useReviewApplication(tournamentId: string, entryId: string) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();
  const mutation = $api.useMutation(
    "post",
    "/api/v1/tournaments/{id}/entries/{entryId}/review"
  );

  function review(
    decision: ReviewDecision,
    options: { reason?: string; onSuccess?: () => void } = {}
  ) {
    const reason = options.reason?.trim();
    mutation.mutate(
      {
        params: { path: { id: tournamentId, entryId } },
        headers: { Authorization: authorizationHeader ?? "" },
        body: reason ? { decision, reason } : { decision },
      },
      {
        onSuccess: () => {
          // The dashboard counts and the entry's status both change.
          void queryClient.invalidateQueries();
          options.onSuccess?.();
        },
      }
    );
  }

  return {
    review,
    isPending: mutation.isPending,
    errorMessage: mutation.isError
      ? extractProblemMessage(
          mutation.error,
          "Something went wrong. Please try again."
        )
      : null,
    reset: mutation.reset,
  };
}
