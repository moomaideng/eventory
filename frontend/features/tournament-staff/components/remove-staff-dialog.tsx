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
import {
  getErrorMessage,
  type StaffMember,
} from "@/features/tournament-staff/utils";

export function RemoveStaffDialog({
  tournamentId,
  staff,
  open,
  onOpenChange,
}: {
  tournamentId: string;
  staff: StaffMember | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const removeMutation = $api.useMutation(
    "delete",
    "/api/v1/tournaments/{id}/staff/{accountId}"
  );

  function handleOpenChange(next: boolean) {
    if (!next) {
      removeMutation.reset();
    }
    onOpenChange(next);
  }

  function handleConfirm() {
    if (!staff) return;

    removeMutation.mutate(
      {
        params: {
          path: {
            id: tournamentId,
            accountId: staff.accountId,
          },
        },
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

  if (!staff) return null;

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Remove staff member</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to remove @{staff.handle}{" "}
            {staff.displayName ? `(${staff.displayName})` : ""} from tournament
            staff? They will lose all staff permissions immediately.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {removeMutation.isError ? (
          <Alert variant="destructive">
            <AlertCircle />
            <AlertTitle>Removal failed</AlertTitle>
            <AlertDescription>
              {getErrorMessage(removeMutation.error)}
            </AlertDescription>
          </Alert>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={removeMutation.isPending}>
            Cancel
          </AlertDialogCancel>
          <AlertDialogAction
            variant="destructive"
            disabled={removeMutation.isPending}
            onClick={handleConfirm}
          >
            {removeMutation.isPending ? (
              <Spinner data-icon="inline-start" />
            ) : null}
            Remove staff
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
