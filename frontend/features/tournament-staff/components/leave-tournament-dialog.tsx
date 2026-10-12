"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { AlertCircle, LogOut } from "lucide-react";
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
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { getErrorMessage } from "@/features/tournament-staff/utils";

export function LeaveTournamentDialog({
  tournamentId,
  tournamentName,
}: {
  tournamentId: string;
  tournamentName?: string;
}) {
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const { user, authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const removeMutation = $api.useMutation(
    "delete",
    "/api/v1/tournaments/{id}/staff/{accountId}"
  );

  function handleOpenChange(next: boolean) {
    if (!next) {
      removeMutation.reset();
    }
    setOpen(next);
  }

  function handleConfirm() {
    if (!user?.id) return;

    removeMutation.mutate(
      {
        params: {
          path: {
            id: tournamentId,
            accountId: user.id,
          },
        },
        headers: { Authorization: authorizationHeader ?? "" },
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries();
          setOpen(false);
          router.push("/organizer");
        },
      }
    );
  }

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogTrigger
        render={
          <Button
            variant="outline"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
          >
            <LogOut data-icon="inline-start" />
            Leave tournament
          </Button>
        }
      />
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Leave tournament</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to leave{" "}
            {tournamentName ? `"${tournamentName}"` : "this tournament"}? You will
            lose all staff access and permissions immediately.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {removeMutation.isError ? (
          <Alert variant="destructive">
            <AlertCircle />
            <AlertTitle>Could not leave tournament</AlertTitle>
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
            disabled={removeMutation.isPending || !user?.id}
            onClick={handleConfirm}
          >
            {removeMutation.isPending ? (
              <Spinner data-icon="inline-start" />
            ) : null}
            Leave tournament
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
