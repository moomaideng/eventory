"use client";

import { useState, type FormEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Settings2 } from "lucide-react";
import { apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import { extractProblemMessage } from "@/lib/schemas";
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
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { fundingSettingsSchema } from "../schemas";

function toLocalInput(value: string) {
  const date = new Date(value);
  return new Date(date.getTime() - date.getTimezoneOffset() * 60_000)
    .toISOString()
    .slice(0, 16);
}

export function FundingSettingsDialog({
  tournamentId,
  goal,
  deadline,
  registrationDeadline,
}: {
  tournamentId: string;
  goal: number;
  deadline: string;
  registrationDeadline: string;
}) {
  const [open, setOpen] = useState(false);
  const [fundingGoal, setFundingGoal] = useState(goal);
  const [fundingDeadline, setFundingDeadline] = useState(
    toLocalInput(deadline)
  );
  const [errors, setErrors] = useState<{
    fundingGoal?: string;
    fundingDeadline?: string;
  }>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  function onOpenChange(next: boolean) {
    if (next) {
      setFundingGoal(goal);
      setFundingDeadline(toLocalInput(deadline));
      setErrors({});
      setApiError(null);
    }
    setOpen(next);
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const parsed = fundingSettingsSchema(registrationDeadline).safeParse({
      fundingGoal,
      fundingDeadline,
    });
    if (!parsed.success) {
      const fields = parsed.error.flatten().fieldErrors;
      setErrors({
        fundingGoal: fields.fundingGoal?.[0],
        fundingDeadline: fields.fundingDeadline?.[0],
      });
      return;
    }
    if (!authorizationHeader) {
      setApiError("Please sign in again.");
      return;
    }
    setErrors({});
    setApiError(null);
    setSaving(true);
    try {
      const { error } = await apiClient.PATCH(
        "/api/v1/tournaments/{tournamentId}",
        {
          params: { path: { tournamentId } },
          headers: { Authorization: authorizationHeader },
          body: {
            fundingGoal: parsed.data.fundingGoal,
            fundingDeadline: new Date(
              parsed.data.fundingDeadline
            ).toISOString(),
          },
        }
      );
      if (error) {
        setApiError(
          extractProblemMessage(error, "Could not update funding details.")
        );
        return;
      }
      await queryClient.invalidateQueries();
      setOpen(false);
    } catch {
      setApiError("Connection failure. Please try again.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger render={<Button variant="outline" size="sm" />}>
        <Settings2 data-icon="inline-start" /> Edit funding
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit crowdfunding</DialogTitle>
          <DialogDescription>
            Update the goal and deadline while funding is active.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <FieldGroup>
            <Field data-invalid={Boolean(errors.fundingGoal)}>
              <FieldLabel htmlFor="edit-funding-goal">
                Funding goal (THB)
              </FieldLabel>
              <Input
                id="edit-funding-goal"
                type="number"
                min={1}
                step={1}
                value={fundingGoal || ""}
                onChange={(event) => setFundingGoal(Number(event.target.value))}
                aria-invalid={Boolean(errors.fundingGoal)}
              />
              {errors.fundingGoal ? (
                <FieldError>{errors.fundingGoal}</FieldError>
              ) : null}
            </Field>
            <Field data-invalid={Boolean(errors.fundingDeadline)}>
              <FieldLabel htmlFor="edit-funding-deadline">
                Funding deadline
              </FieldLabel>
              <Input
                id="edit-funding-deadline"
                type="datetime-local"
                value={fundingDeadline}
                onChange={(event) => setFundingDeadline(event.target.value)}
                aria-invalid={Boolean(errors.fundingDeadline)}
              />
              {errors.fundingDeadline ? (
                <FieldError>{errors.fundingDeadline}</FieldError>
              ) : null}
            </Field>
          </FieldGroup>
          {apiError ? (
            <Alert variant="destructive">
              <AlertTitle>Could not save</AlertTitle>
              <AlertDescription>{apiError}</AlertDescription>
            </Alert>
          ) : null}
          <DialogFooter>
            <DialogClose
              render={<Button type="button" variant="outline" />}
              disabled={saving}
            >
              Cancel
            </DialogClose>
            <Button type="submit" disabled={saving}>
              {saving ? <Spinner data-icon="inline-start" /> : null}Save changes
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
