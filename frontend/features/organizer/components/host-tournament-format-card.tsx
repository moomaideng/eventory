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
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";

export interface HostTournamentFormatCardProps {
  registrationMode: "SOLO" | "TEAM";
  onModeChange: (mode: "SOLO" | "TEAM") => void;
  minTeamSize: number;
  onMinTeamSizeChange: (value: number) => void;
  maxTeamSize: number;
  onMaxTeamSizeChange: (value: number) => void;
  capacity: number;
  onCapacityChange: (value: number) => void;
  entryFee: number;
  onEntryFeeChange: (value: number) => void;
  errors: Partial<
    Record<
      | "registrationMode"
      | "minTeamSize"
      | "maxTeamSize"
      | "capacity"
      | "entryFee",
      string
    >
  >;
}

export function HostTournamentFormatCard({
  registrationMode,
  onModeChange,
  minTeamSize,
  onMinTeamSizeChange,
  maxTeamSize,
  onMaxTeamSizeChange,
  capacity,
  onCapacityChange,
  entryFee,
  onEntryFeeChange,
  errors,
}: HostTournamentFormatCardProps) {
  return (
    <Card className="border-border">
      <CardHeader>
        <CardTitle className="text-lg">Format & Capacity</CardTitle>
        <CardDescription>
          Configure participation mode, roster boundaries, maximum participant capacity, and entry fees.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <FieldGroup>
          <Field>
            <FieldLabel>Registration Mode</FieldLabel>
            <RadioGroup
              value={registrationMode}
              onValueChange={(val) => onModeChange(val as "SOLO" | "TEAM")}
              className="flex flex-row gap-4 pt-1"
            >
              <FieldLabel
                htmlFor="mode-team"
                className="flex flex-1 cursor-pointer items-center justify-between rounded-lg border p-3.5 transition-colors has-data-checked:border-primary has-data-checked:bg-primary/5"
              >
                <div className="flex flex-col gap-0.5">
                  <span className="text-sm font-semibold">Team Tournament</span>
                  <span className="text-muted-foreground text-xs font-normal">
                    Competitors register as squads
                  </span>
                </div>
                <RadioGroupItem value="TEAM" id="mode-team" />
              </FieldLabel>

              <FieldLabel
                htmlFor="mode-solo"
                className="flex flex-1 cursor-pointer items-center justify-between rounded-lg border p-3.5 transition-colors has-data-checked:border-primary has-data-checked:bg-primary/5"
              >
                <div className="flex flex-col gap-0.5">
                  <span className="text-sm font-semibold">Solo Tournament</span>
                  <span className="text-muted-foreground text-xs font-normal">
                    1v1 individual participants
                  </span>
                </div>
                <RadioGroupItem value="SOLO" id="mode-solo" />
              </FieldLabel>
            </RadioGroup>
          </Field>

          {registrationMode === "TEAM" ? (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <Field data-invalid={Boolean(errors.minTeamSize)}>
                <FieldLabel htmlFor="min-team-size">Min Team Size</FieldLabel>
                <Input
                  id="min-team-size"
                  type="number"
                  min={1}
                  max={20}
                  value={minTeamSize}
                  onChange={(e) =>
                    onMinTeamSizeChange(
                      Math.max(1, parseInt(e.target.value, 10) || 1)
                    )
                  }
                  aria-invalid={Boolean(errors.minTeamSize)}
                />
                {errors.minTeamSize ? (
                  <FieldError>{errors.minTeamSize}</FieldError>
                ) : null}
              </Field>

              <Field data-invalid={Boolean(errors.maxTeamSize)}>
                <FieldLabel htmlFor="max-team-size">Max Team Size</FieldLabel>
                <Input
                  id="max-team-size"
                  type="number"
                  min={minTeamSize}
                  max={20}
                  value={maxTeamSize}
                  onChange={(e) =>
                    onMaxTeamSizeChange(
                      Math.max(1, parseInt(e.target.value, 10) || 1)
                    )
                  }
                  aria-invalid={Boolean(errors.maxTeamSize)}
                />
                {errors.maxTeamSize ? (
                  <FieldError>{errors.maxTeamSize}</FieldError>
                ) : null}
              </Field>
            </div>
          ) : null}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field data-invalid={Boolean(errors.capacity)}>
              <FieldLabel htmlFor="tournament-capacity">
                {registrationMode === "SOLO"
                  ? "Participant Capacity"
                  : "Team Capacity"}
              </FieldLabel>
              <Input
                id="tournament-capacity"
                type="number"
                min={1}
                max={256}
                value={capacity}
                onChange={(e) =>
                  onCapacityChange(
                    Math.max(1, parseInt(e.target.value, 10) || 1)
                  )
                }
                aria-invalid={Boolean(errors.capacity)}
              />
              <FieldDescription>
                Maximum number of accepted{" "}
                {registrationMode === "SOLO" ? "participants" : "teams"}.
              </FieldDescription>
              {errors.capacity ? (
                <FieldError>{errors.capacity}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(errors.entryFee)}>
              <FieldLabel htmlFor="entry-fee">Entry Fee (THB)</FieldLabel>
              <Input
                id="entry-fee"
                type="number"
                min={0}
                value={entryFee}
                onChange={(e) =>
                  onEntryFeeChange(
                    Math.max(0, parseInt(e.target.value, 10) || 0)
                  )
                }
                aria-invalid={Boolean(errors.entryFee)}
              />
              <FieldDescription>Set to 0 for free entry.</FieldDescription>
              {errors.entryFee ? (
                <FieldError>{errors.entryFee}</FieldError>
              ) : null}
            </Field>
          </div>
        </FieldGroup>
      </CardContent>
    </Card>
  );
}
