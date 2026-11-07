"use client";

import { useState } from "react";
import { CalendarClock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { matchScheduleSchema } from "../schemas";
import type { MatchResponse } from "../utils";

// ISO string -> "YYYY-MM-DDTHH:mm" in local time, as <input type="datetime-local"> expects.
function toLocalInput(iso?: string) {
  if (!iso) return "";
  const d = new Date(iso);
  return new Date(d.getTime() - d.getTimezoneOffset() * 60_000).toISOString().slice(0, 16);
}

interface MatchScheduleFormProps {
  match: MatchResponse;
  onSchedule: (matchId: string, scheduledAt: string, venue: string) => Promise<void>;
  isScheduling: boolean;
  onClose: () => void;
}

export function MatchScheduleForm({ match, onSchedule, isScheduling, onClose }: MatchScheduleFormProps) {
  const [scheduledAt, setScheduledAt] = useState(toLocalInput(match.scheduledAt));
  const [venue, setVenue] = useState(match.venue);
  const [errors, setErrors] = useState<{ scheduledAt?: string; venue?: string; form?: string }>({});

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const parsed = matchScheduleSchema.safeParse({ scheduledAt, venue });
    if (!parsed.success) {
      const issues = parsed.error.flatten().fieldErrors;
      setErrors({ scheduledAt: issues.scheduledAt?.[0], venue: issues.venue?.[0] });
      return;
    }

    try {
      setErrors({});
      await onSchedule(match.id, new Date(parsed.data.scheduledAt).toISOString(), parsed.data.venue);
      onClose();
    } catch (err: unknown) {
      setErrors({ form: err instanceof Error ? err.message : "Failed to save schedule." });
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      {errors.form ? (
        <div className="rounded-md bg-destructive/15 p-3 text-sm text-destructive font-medium">
          {errors.form}
        </div>
      ) : null}

      <FieldGroup>
        <Field data-invalid={Boolean(errors.scheduledAt)}>
          <FieldLabel htmlFor="match-scheduled-at">Start time</FieldLabel>
          <Input
            id="match-scheduled-at"
            type="datetime-local"
            value={scheduledAt}
            onChange={(e) => setScheduledAt(e.target.value)}
            aria-invalid={Boolean(errors.scheduledAt)}
          />
          {errors.scheduledAt ? <FieldError>{errors.scheduledAt}</FieldError> : null}
        </Field>

        <Field data-invalid={Boolean(errors.venue)}>
          <FieldLabel htmlFor="match-schedule-venue">Venue</FieldLabel>
          <Input
            id="match-schedule-venue"
            value={venue}
            onChange={(e) => setVenue(e.target.value)}
            placeholder="e.g. Court 1, or a Discord link"
            aria-invalid={Boolean(errors.venue)}
          />
          {errors.venue ? <FieldError>{errors.venue}</FieldError> : null}
        </Field>
      </FieldGroup>

      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" size="sm" onClick={onClose}>
          Close
        </Button>
        <Button type="submit" size="sm" disabled={isScheduling}>
          {isScheduling ? <Spinner /> : <CalendarClock data-icon="inline-start" />}
          Save Schedule
        </Button>
      </div>
    </form>
  );
}
