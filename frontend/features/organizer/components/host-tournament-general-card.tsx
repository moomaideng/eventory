"use client";

import React from "react";
import { Globe } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

export interface HostTournamentGeneralCardProps {
  organizerName?: string;
  name: string;
  onNameChange: (value: string) => void;
  game: string;
  onGameChange: (value: string) => void;
  location: string;
  onLocationChange: (value: string) => void;
  description: string;
  onDescriptionChange: (value: string) => void;
  errors: Partial<Record<"name" | "game" | "location" | "description", string>>;
}

export function HostTournamentGeneralCard({
  organizerName,
  name,
  onNameChange,
  game,
  onGameChange,
  location,
  onLocationChange,
  description,
  onDescriptionChange,
  errors,
}: HostTournamentGeneralCardProps) {
  return (
    <Card className="border-border">
      <CardHeader>
        <CardTitle className="text-lg">General Information</CardTitle>
        <CardDescription>
          {organizerName
            ? `Set tournament identity and rules published under ${organizerName}.`
            : "Provide the tournament title, game discipline, venue or server location, and competition rules."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <FieldGroup>
          <Field data-invalid={Boolean(errors.name)}>
            <FieldLabel htmlFor="tournament-name">Tournament Name</FieldLabel>
            <Input
              id="tournament-name"
              placeholder="e.g. Bangkok Invitational Season 1"
              maxLength={160}
              value={name}
              onChange={(e) => onNameChange(e.target.value)}
              aria-invalid={Boolean(errors.name)}
            />
            {errors.name ? <FieldError>{errors.name}</FieldError> : null}
          </Field>

          <Field data-invalid={Boolean(errors.game)}>
            <FieldLabel htmlFor="tournament-game">Game Title</FieldLabel>
            <Input
              id="tournament-game"
              placeholder="e.g. Valorant, League of Legends, Tekken 8, Counter-Strike 2"
              maxLength={80}
              value={game}
              onChange={(e) => onGameChange(e.target.value)}
              aria-invalid={Boolean(errors.game)}
            />
            {errors.game ? <FieldError>{errors.game}</FieldError> : null}
          </Field>

          <Field data-invalid={Boolean(errors.location)}>
            <div className="flex items-center justify-between">
              <FieldLabel htmlFor="tournament-location">Location</FieldLabel>
              <Button
                type="button"
                variant="outline"
                size="xs"
                onClick={() => onLocationChange("Online")}
                className="text-muted-foreground hover:text-foreground"
              >
                <Globe data-icon="inline-start" />
                Set as Online
              </Button>
            </div>
            <Input
              id="tournament-location"
              placeholder="e.g. Online, Discord Server, or Venue Address"
              maxLength={160}
              value={location}
              onChange={(e) => onLocationChange(e.target.value)}
              aria-invalid={Boolean(errors.location)}
            />
            {errors.location ? (
              <FieldError>{errors.location}</FieldError>
            ) : null}
          </Field>

          <Field data-invalid={Boolean(errors.description)}>
            <FieldLabel htmlFor="tournament-desc">
              Tournament Description & Rules
            </FieldLabel>
            <Textarea
              id="tournament-desc"
              placeholder="Outline match formats, map pools, bracket rules, check-in instructions, and communication channels..."
              rows={4}
              maxLength={4000}
              value={description}
              onChange={(e) => onDescriptionChange(e.target.value)}
              aria-invalid={Boolean(errors.description)}
            />
            {errors.description ? (
              <FieldError>{errors.description}</FieldError>
            ) : null}
          </Field>
        </FieldGroup>
      </CardContent>
    </Card>
  );
}
