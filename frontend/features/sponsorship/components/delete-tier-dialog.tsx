"use client";

import { useQueryClient } from "@tanstack/react-query";
import { AlertCircle } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
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
import { Spinner } from "@/components/ui/spinner";
import { type SponsorshipTier } from "@/features/sponsorship/utils";
import { getErrorMessage } from "@/features/tournament-staff/utils";

export function DeleteTierDialog({
  tournamentId,
  tier,
  open,
  onOpenChange,
}: {
  tournamentId: string;
  tier: SponsorshipTier | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();
  const deleteMutation = $api.useMutation(
    "delete",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers/{tierId}"
  );

  function handleOpenChange(next: boolean) {
    if (!next) deleteMutation.reset();
    onOpenChange(next);
  }

  function handleConfirm() {
    if (!tier) return;
    deleteMutation.mutate(
      {
        params: { path: { tournamentId, tierId: tier.id } },
        headers: { Authorization: authorizationHeader ?? "" },
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries();
          onOpenChange(false);
        },
      }
    );
  }

  if (!tier) return null;

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Archive {tier.name} tier</AlertDialogTitle>
          <AlertDialogDescription>
            Sponsors will no longer see or purchase this package. Existing
            checkout and payment history will be kept.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {deleteMutation.isError ? (
          <Alert variant="destructive">
            <AlertCircle />
            <AlertTitle>Archive failed</AlertTitle>
            <AlertDescription>
              {getErrorMessage(deleteMutation.error)}
            </AlertDescription>
          </Alert>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={deleteMutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={deleteMutation.isPending}
            onClick={handleConfirm}
          >
            {deleteMutation.isPending ? (
              <Spinner data-icon="inline-start" />
            ) : null}
            Archive tier
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
