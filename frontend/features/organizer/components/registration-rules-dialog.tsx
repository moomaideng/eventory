"use client";

import { useState, type FormEvent } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Pencil } from "lucide-react";
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
import { Spinner } from "@/components/ui/spinner";
import { HostTournamentFormatCard } from "@/features/organizer/components/host-tournament-format-card";
import { registrationRulesSchema } from "@/features/organizer/schemas";

type RegistrationRules = {
  registrationMode: "SOLO" | "TEAM";
  minTeamSize: number;
  maxTeamSize: number;
  capacity: number;
  entryFee: number;
};

export function RegistrationRulesDialog({
  tournamentId,
  rules,
}: {
  tournamentId: string;
  rules: RegistrationRules;
}) {
  const [open, setOpen] = useState(false);
  const [mode, setMode] = useState(rules.registrationMode);
  const [minTeamSize, setMinTeamSize] = useState(rules.minTeamSize);
  const [maxTeamSize, setMaxTeamSize] = useState(rules.maxTeamSize);
  const [capacity, setCapacity] = useState(rules.capacity);
  const [entryFee, setEntryFee] = useState(rules.entryFee);
  const [errors, setErrors] = useState<{
    registrationMode?: string;
    minTeamSize?: string;
    maxTeamSize?: string;
    capacity?: string;
    entryFee?: string;
  }>({});
  const [apiError, setApiError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  function onOpenChange(next: boolean) {
    if (next) {
      setMode(rules.registrationMode);
      setMinTeamSize(rules.minTeamSize);
      setMaxTeamSize(rules.maxTeamSize);
      setCapacity(rules.capacity);
      setEntryFee(rules.entryFee);
      setErrors({});
      setApiError(null);
    }
    setOpen(next);
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const parsed = registrationRulesSchema.safeParse({
      registrationMode: mode,
      minTeamSize,
      maxTeamSize,
      capacity,
      entryFee,
    });
    if (!parsed.success) {
      const fields = parsed.error.flatten().fieldErrors;
      setErrors({
        registrationMode: fields.registrationMode?.[0],
        minTeamSize: fields.minTeamSize?.[0],
        maxTeamSize: fields.maxTeamSize?.[0],
        capacity: fields.capacity?.[0],
        entryFee: fields.entryFee?.[0],
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
            registrationMode: parsed.data.registrationMode,
            minTeamSize: parsed.data.minTeamSize,
            maxTeamSize: parsed.data.maxTeamSize,
            capacity: parsed.data.capacity,
            entryFee: parsed.data.entryFee,
          },
        }
      );
      if (error) {
        setApiError(
          extractProblemMessage(error, "Could not update registration rules.")
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
        <Pencil data-icon="inline-start" /> Edit registration rules
      </DialogTrigger>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Edit registration rules</DialogTitle>
          <DialogDescription>
            Review these rules before you open participant registration.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="flex flex-col gap-5">
          <HostTournamentFormatCard
            registrationMode={mode}
            onModeChange={setMode}
            minTeamSize={minTeamSize}
            onMinTeamSizeChange={setMinTeamSize}
            maxTeamSize={maxTeamSize}
            onMaxTeamSizeChange={setMaxTeamSize}
            capacity={capacity}
            onCapacityChange={setCapacity}
            entryFee={entryFee}
            onEntryFeeChange={setEntryFee}
            errors={errors}
          />
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
              {saving ? <Spinner data-icon="inline-start" /> : null}
              Save rules
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
