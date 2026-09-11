"use client";

import { useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, History, Settings2 } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
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
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";
import {
  REASON_MAX_LENGTH,
  REASON_MIN_LENGTH,
  statusLabel,
  statusOverrideSchema,
  TERMINAL_WARNING,
  type TournamentStatus,
} from "@/features/organizer/utils";

export function StatusControlDialog({
  tournamentId,
  currentStatus,
}: {
  tournamentId: string;
  currentStatus: string;
}) {
  const [open, setOpen] = useState(false);
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const [status, setStatus] = useState<string | null>(null);
  const [reason, setReason] = useState("");
  const [submitted, setSubmitted] = useState(false);

  // The server owns the transition table; the dialog renders whatever it is
  // told is reachable rather than keeping its own copy of the rules.
  const history = $api.useQuery(
    "get",
    "/api/v1/tournaments/{id}/status/history",
    {
      params: { path: { id: tournamentId } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { enabled: open && Boolean(authorizationHeader), retry: false }
  );

  const override = $api.useMutation("patch", "/api/v1/tournaments/{id}/status");

  // Reset on open rather than in an effect, so a previous attempt — including a
  // failed one — never leaks into the next without triggering a cascading
  // render on every open.
  function handleOpenChange(next: boolean) {
    if (next) {
      setStatus(null);
      setReason("");
      setSubmitted(false);
      override.reset();
    }
    setOpen(next);
  }

  const isTerminal =
    history.isSuccess && (history.data?.allowedTransitions?.length ?? 0) === 0;

  const options = useMemo(
    () =>
      (history.data?.allowedTransitions ?? []).map((value) => ({
        value,
        label: statusLabel(value),
      })),
    [history.data?.allowedTransitions]
  );

  const parsed = statusOverrideSchema.safeParse({ status, reason });
  const fieldErrors = submitted && !parsed.success
    ? parsed.error.flatten().fieldErrors
    : {};

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitted(true);
    if (!parsed.success) return;

    override.mutate(
      {
        params: { path: { id: tournamentId } },
        headers: { Authorization: authorizationHeader ?? "" },
        // The value came from the server's own allowedTransitions list, so it
        // is one of the schema's statuses; zod only checks that one was picked.
        body: {
          ...parsed.data,
          status: parsed.data.status as TournamentStatus,
        },
      },
      {
        onSuccess: () => {
          // The dashboard, the public page and this dialog all read the status.
          void queryClient.invalidateQueries();
          setOpen(false);
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={<Button variant="outline" />}>
        <Settings2 data-icon="inline-start" />
        Change status
      </DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Change tournament status</DialogTitle>
          <DialogDescription>
            This tournament is currently{" "}
            <span className="font-medium capitalize">
              {statusLabel(currentStatus)}
            </span>
            . The change is recorded with your name, the reason, and the time.
          </DialogDescription>
        </DialogHeader>

        {history.isLoading ? (
          <div className="flex flex-col gap-3">
            <Skeleton className="h-9 w-full" />
            <Skeleton className="h-24 w-full" />
          </div>
        ) : history.isError ? (
          <Alert variant="destructive">
            <AlertTriangle />
            <AlertTitle>Could not load status options</AlertTitle>
            <AlertDescription>
              Close this dialog and try again.
            </AlertDescription>
          </Alert>
        ) : isTerminal ? (
          <Alert>
            <AlertTriangle />
            <AlertTitle>This status is final</AlertTitle>
            <AlertDescription>
              A {statusLabel(currentStatus)} tournament cannot be moved to
              another status.
            </AlertDescription>
          </Alert>
        ) : (
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <FieldGroup>
              <Field data-invalid={fieldErrors.status ? true : undefined}>
                <FieldLabel htmlFor="status-override">New status</FieldLabel>
                <Select
                  items={options}
                  value={status}
                  onValueChange={(value) => setStatus(value as string | null)}
                >
                  <SelectTrigger
                    id="status-override"
                    aria-invalid={fieldErrors.status ? true : undefined}
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      {options.map((option) => (
                        <SelectItem
                          key={option.value}
                          value={option.value}
                          className="capitalize"
                        >
                          {option.label}
                        </SelectItem>
                      ))}
                    </SelectGroup>
                  </SelectContent>
                </Select>
                {fieldErrors.status ? (
                  <FieldError>{fieldErrors.status[0]}</FieldError>
                ) : null}
              </Field>

              <Field data-invalid={fieldErrors.reason ? true : undefined}>
                <FieldLabel htmlFor="status-reason">Reason</FieldLabel>
                <Textarea
                  id="status-reason"
                  rows={4}
                  value={reason}
                  maxLength={REASON_MAX_LENGTH}
                  aria-invalid={fieldErrors.reason ? true : undefined}
                  onChange={(event) => setReason(event.target.value)}
                  placeholder="Why is the status changing?"
                />
                {fieldErrors.reason ? (
                  <FieldError>{fieldErrors.reason[0]}</FieldError>
                ) : (
                  <FieldDescription>
                    {reason.trim().length}/{REASON_MAX_LENGTH} characters, at
                    least {REASON_MIN_LENGTH}. Everyone reviewing this
                    tournament can see it.
                  </FieldDescription>
                )}
              </Field>
            </FieldGroup>

            {status && TERMINAL_WARNING[status] ? (
              <Alert variant="destructive">
                <AlertTriangle />
                <AlertTitle>This cannot be undone</AlertTitle>
                <AlertDescription>{TERMINAL_WARNING[status]}</AlertDescription>
              </Alert>
            ) : null}

            {override.isError ? (
              <Alert variant="destructive">
                <AlertTriangle />
                <AlertTitle>The status was not changed</AlertTitle>
                <AlertDescription>
                  {overrideErrorMessage(override.error)}
                </AlertDescription>
              </Alert>
            ) : null}

            <DialogFooter>
              <DialogClose
                render={<Button variant="outline" type="button" />}
                disabled={override.isPending}
              >
                Cancel
              </DialogClose>
              <Button type="submit" disabled={override.isPending}>
                {override.isPending ? (
                  <Spinner data-icon="inline-start" />
                ) : null}
                Change status
              </Button>
            </DialogFooter>
          </form>
        )}

        {history.data?.items?.length ? (
          <>
            <Separator />
            <div className="flex flex-col gap-3">
              <p className="text-muted-foreground flex items-center gap-2 text-sm font-medium">
                <History aria-hidden="true" className="size-4" />
                Previous changes
              </p>
              <ul className="flex max-h-40 flex-col gap-3 overflow-y-auto">
                {history.data.items.map((item) => (
                  <li key={item.id} className="flex flex-col gap-1 text-sm">
                    <span className="capitalize">
                      {statusLabel(item.fromStatus)} →{" "}
                      {statusLabel(item.toStatus)}
                    </span>
                    <span className="text-muted-foreground wrap-anywhere">
                      {item.reason}
                    </span>
                    <span className="text-muted-foreground text-xs">
                      {item.actorName ?? "Unknown"} ·{" "}
                      {new Date(item.createdAt).toLocaleString()}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          </>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}

// The API reports refusals as a plain detail string; anything else is a network
// or server fault the organizer cannot act on.
function overrideErrorMessage(error: unknown): string {
  if (
    error &&
    typeof error === "object" &&
    "detail" in error &&
    typeof (error as { detail: unknown }).detail === "string"
  ) {
    return (error as { detail: string }).detail;
  }
  return "Something went wrong. Please try again.";
}
