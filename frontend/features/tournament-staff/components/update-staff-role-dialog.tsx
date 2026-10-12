"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertCircle } from "lucide-react";
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
} from "@/components/ui/dialog";
import {
  Field,
  FieldDescription,
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
import { Spinner } from "@/components/ui/spinner";
import {
  getErrorMessage,
  ROLE_OPTIONS,
  type StaffMember,
} from "@/features/tournament-staff/utils";

function UpdateStaffRoleForm({
  tournamentId,
  staff,
  onSuccess,
}: {
  tournamentId: string;
  staff: StaffMember;
  onSuccess: () => void;
}) {
  const [role, setRole] = useState<"ADMIN" | "REFEREE">(
    staff.role === "ADMIN" ? "ADMIN" : "REFEREE"
  );
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const updateMutation = $api.useMutation(
    "patch",
    "/api/v1/tournaments/{id}/staff/{accountId}"
  );

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();

    updateMutation.mutate(
      {
        params: {
          path: {
            id: tournamentId,
            accountId: staff.accountId,
          },
        },
        headers: { Authorization: authorizationHeader ?? "" },
        body: { role },
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries();
          onSuccess();
        },
      }
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6">
      <DialogHeader>
        <DialogTitle>Change staff role</DialogTitle>
        <DialogDescription>
          Update the staff role for @{staff.handle}{" "}
          {staff.displayName ? `(${staff.displayName})` : ""}.
        </DialogDescription>
      </DialogHeader>

      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="update-role">New staff role</FieldLabel>
          <Select
            items={ROLE_OPTIONS}
            value={role}
            onValueChange={(val) => setRole(val as "ADMIN" | "REFEREE")}
          >
            <SelectTrigger id="update-role" className="w-full">
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
          <FieldDescription>
            Assign either Admin or Referee privileges.
          </FieldDescription>
        </Field>
      </FieldGroup>

      {updateMutation.isError ? (
        <Alert variant="destructive">
          <AlertCircle />
          <AlertTitle>Update failed</AlertTitle>
          <AlertDescription>
            {getErrorMessage(updateMutation.error)}
          </AlertDescription>
        </Alert>
      ) : null}

      <DialogFooter>
        <DialogClose
          render={<Button variant="outline" type="button" />}
          disabled={updateMutation.isPending}
        >
          Cancel
        </DialogClose>
        <Button type="submit" disabled={updateMutation.isPending}>
          {updateMutation.isPending ? (
            <Spinner data-icon="inline-start" />
          ) : null}
          Save changes
        </Button>
      </DialogFooter>
    </form>
  );
}

export function UpdateStaffRoleDialog({
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
  if (!staff) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <UpdateStaffRoleForm
          key={staff.accountId}
          tournamentId={tournamentId}
          staff={staff}
          onSuccess={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
