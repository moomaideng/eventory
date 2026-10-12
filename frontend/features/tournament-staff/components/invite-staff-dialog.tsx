"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertCircle, UserPlus } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import {
  inviteStaffSchema,
  type InviteStaffInput,
} from "@/features/tournament-staff/schemas";
import {
  getErrorMessage,
  ROLE_OPTIONS,
} from "@/features/tournament-staff/utils";

export function InviteStaffDialog({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const [open, setOpen] = useState(false);
  const [handle, setHandle] = useState("");
  const [role, setRole] = useState<"ADMIN" | "REFEREE">("REFEREE");
  const [submitted, setSubmitted] = useState(false);

  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const inviteMutation = $api.useMutation(
    "post",
    "/api/v1/tournaments/{id}/staff"
  );

  function handleOpenChange(next: boolean) {
    if (next) {
      setHandle("");
      setRole("REFEREE");
      setSubmitted(false);
      inviteMutation.reset();
    }
    setOpen(next);
  }

  const parsed = inviteStaffSchema.safeParse({
    handle: handle.trim(),
    role,
  });

  const fieldErrors =
    submitted && !parsed.success ? parsed.error.flatten().fieldErrors : {};

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitted(true);
    if (!parsed.success) return;

    inviteMutation.mutate(
      {
        params: { path: { id: tournamentId } },
        headers: { Authorization: authorizationHeader ?? "" },
        body: parsed.data as InviteStaffInput,
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries();
          setOpen(false);
        },
      }
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={<Button size="sm" />}>
        <UserPlus data-icon="inline-start" />
        Invite staff
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          <DialogHeader>
            <DialogTitle>Invite staff member</DialogTitle>
            <DialogDescription>
              Invite an active user by their handle to help run this tournament.
            </DialogDescription>
          </DialogHeader>

          <FieldGroup>
            <Field data-invalid={fieldErrors.handle ? true : undefined}>
              <FieldLabel htmlFor="staff-handle">User handle</FieldLabel>
              <Input
                id="staff-handle"
                placeholder="username"
                value={handle}
                onChange={(e) => setHandle(e.target.value)}
                aria-invalid={fieldErrors.handle ? true : undefined}
                autoComplete="off"
              />
              {fieldErrors.handle ? (
                <FieldError>{fieldErrors.handle[0]}</FieldError>
              ) : (
                <FieldDescription>
                  Enter the handle without the &apos;@&apos; prefix.
                </FieldDescription>
              )}
            </Field>

            <Field data-invalid={fieldErrors.role ? true : undefined}>
              <FieldLabel htmlFor="staff-role">Staff role</FieldLabel>
              <Select
                items={ROLE_OPTIONS}
                value={role}
                onValueChange={(val) => setRole(val as "ADMIN" | "REFEREE")}
              >
                <SelectTrigger
                  id="staff-role"
                  aria-invalid={fieldErrors.role ? true : undefined}
                  className="w-full"
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    {ROLE_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
              {fieldErrors.role ? (
                <FieldError>{fieldErrors.role[0]}</FieldError>
              ) : (
                <FieldDescription>
                  Admins and referees can manage operational workflows.
                </FieldDescription>
              )}
            </Field>
          </FieldGroup>

          {inviteMutation.isError ? (
            <Alert variant="destructive">
              <AlertCircle />
              <AlertTitle>Invitation failed</AlertTitle>
              <AlertDescription>
                {getErrorMessage(inviteMutation.error)}
              </AlertDescription>
            </Alert>
          ) : null}

          <DialogFooter>
            <DialogClose
              render={<Button variant="outline" type="button" />}
              disabled={inviteMutation.isPending}
            >
              Cancel
            </DialogClose>
            <Button type="submit" disabled={inviteMutation.isPending}>
              {inviteMutation.isPending ? (
                <Spinner data-icon="inline-start" />
              ) : null}
              Send invite
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
