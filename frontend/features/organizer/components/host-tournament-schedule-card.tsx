"use client";

import React from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";

export interface HostTournamentScheduleCardProps {
  registrationDeadline: string;
  onRegistrationDeadlineChange: (value: string) => void;
  startAt: string;
  onStartAtChange: (value: string) => void;
  endAt: string;
  onEndAtChange: (value: string) => void;
  errors: Partial<
    Record<"registrationDeadline" | "startAt" | "endAt", string>
  >;
}

export function HostTournamentScheduleCard({
  registrationDeadline,
  onRegistrationDeadlineChange,
  startAt,
  onStartAtChange,
  endAt,
  onEndAtChange,
  errors,
}: HostTournamentScheduleCardProps) {
  return (
    <Card className="border-border">
      <CardHeader>
        <CardTitle className="text-lg">Tournament Schedule</CardTitle>
        <CardDescription>
          Specify registration closing and the tournament running timeframe. Invariant: Registration Deadline &lt; Start Time &lt; End Time.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <FieldGroup>
          <Field data-invalid={Boolean(errors.registrationDeadline)}>
            <FieldLabel htmlFor="registration-deadline">
              Registration Deadline
            </FieldLabel>
            <Input
              id="registration-deadline"
              type="datetime-local"
              value={registrationDeadline}
              onChange={(e) => onRegistrationDeadlineChange(e.target.value)}
              aria-invalid={Boolean(errors.registrationDeadline)}
            />
            {errors.registrationDeadline ? (
              <FieldError>{errors.registrationDeadline}</FieldError>
            ) : (
              <FieldDescription>
                When participant and squad registrations close.
              </FieldDescription>
            )}
          </Field>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={Boolean(errors.startAt)}>
              <FieldLabel htmlFor="tournament-start">
                Tournament Start
              </FieldLabel>
              <Input
                id="tournament-start"
                type="datetime-local"
                value={startAt}
                onChange={(e) => onStartAtChange(e.target.value)}
                aria-invalid={Boolean(errors.startAt)}
              />
              {errors.startAt ? (
                <FieldError>{errors.startAt}</FieldError>
              ) : (
                <FieldDescription>Must be after deadline.</FieldDescription>
              )}
            </Field>

            <Field data-invalid={Boolean(errors.endAt)}>
              <FieldLabel htmlFor="tournament-end">Tournament End</FieldLabel>
              <Input
                id="tournament-end"
                type="datetime-local"
                value={endAt}
                onChange={(e) => onEndAtChange(e.target.value)}
                aria-invalid={Boolean(errors.endAt)}
              />
              {errors.endAt ? (
                <FieldError>{errors.endAt}</FieldError>
              ) : (
                <FieldDescription>Must be after start.</FieldDescription>
              )}
            </Field>
          </div>
        </FieldGroup>
      </CardContent>
    </Card>
  );
}
