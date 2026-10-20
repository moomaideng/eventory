"use client";

import { useState } from "react";
import { AlertCircle, Check } from "lucide-react";
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
import { useReviewApplication } from "@/features/organizer/hooks/use-review-application";

export function AcceptApplicationDialog({
  tournamentId,
  teamId,
  teamName,
}: {
  tournamentId: string;
  teamId: string;
  teamName: string;
}) {
  const [open, setOpen] = useState(false);
  const { review, isPending, errorMessage, reset } = useReviewApplication(
    tournamentId,
    teamId
  );

  function handleOpenChange(next: boolean) {
    // Keep the dialog up while the request is in flight so its result is seen.
    if (isPending) return;
    if (next) reset();
    setOpen(next);
  }

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogTrigger
        render={
          <Button
            variant="outline"
            size="sm"
            aria-label={`Accept ${teamName}`}
          />
        }
      >
        <Check data-icon="inline-start" />
        Accept
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Accept this application?</AlertDialogTitle>
          <AlertDialogDescription>
            <span className="font-medium">{teamName}</span> will be confirmed
            for a spot in this tournament. This decision is final.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {errorMessage ? (
          <Alert variant="destructive">
            <AlertCircle />
            <AlertTitle>The application was not accepted</AlertTitle>
            <AlertDescription>{errorMessage}</AlertDescription>
          </Alert>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isPending}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            disabled={isPending}
            onClick={() =>
              review("ACCEPT", { onSuccess: () => setOpen(false) })
            }
          >
            {isPending ? <Spinner data-icon="inline-start" /> : null}
            Accept
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
