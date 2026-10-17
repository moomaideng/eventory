"use client";

import { useState } from "react";
import { AlertCircle, Info, X } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import { useReviewApplication } from "@/features/organizer/hooks/use-review-application";
import { REASON_MAX_LENGTH } from "@/features/organizer/utils";

export function RejectApplicationDialog({
  tournamentId,
  teamId,
  teamName,
}: {
  tournamentId: string;
  teamId: string;
  teamName: string;
}) {
  const [open, setOpen] = useState(false);
  const [reason, setReason] = useState("");
  const { review, isPending, errorMessage, reset } = useReviewApplication(
    tournamentId,
    teamId
  );

  function handleOpenChange(next: boolean) {
    // Keep the dialog up while the request is in flight so its result is seen.
    if (isPending) return;
    if (next) {
      setReason("");
      reset();
    }
    setOpen(next);
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    review("REJECT", { reason, onSuccess: () => setOpen(false) });
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger
        render={
          <Button
            variant="outline"
            size="sm"
            aria-label={`Reject ${teamName}`}
          />
        }
      >
        <X data-icon="inline-start" />
        Reject
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Reject this application?</DialogTitle>
          <DialogDescription>
            <span className="font-medium">{teamName}</span> will be rejected.
            The decision is recorded with your name and the time, and you can
            accept it again later.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor={`reject-reason-${teamId}`}>
                Reason (optional)
              </FieldLabel>
              <Textarea
                id={`reject-reason-${teamId}`}
                rows={4}
                value={reason}
                maxLength={REASON_MAX_LENGTH}
                onChange={(event) => setReason(event.target.value)}
                placeholder="Why is this application being rejected?"
              />
              <FieldDescription>
                {reason.length}/{REASON_MAX_LENGTH} characters. Kept in the
                audit trail.
              </FieldDescription>
            </Field>
          </FieldGroup>

          {/* TODO(US7-2): remove once rejecting refunds the entry fee. */}
          <Alert>
            <Info />
            <AlertTitle>No automatic refund</AlertTitle>
            <AlertDescription>
              Rejecting does not automatically refund any entry fee that was
              paid.
            </AlertDescription>
          </Alert>

          {errorMessage ? (
            <Alert variant="destructive">
              <AlertCircle />
              <AlertTitle>The application was not rejected</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </Alert>
          ) : null}

          <DialogFooter>
            <DialogClose
              render={<Button variant="outline" type="button" />}
              disabled={isPending}
            >
              Cancel
            </DialogClose>
            <Button type="submit" variant="destructive" disabled={isPending}>
              {isPending ? <Spinner data-icon="inline-start" /> : null}
              Reject
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
