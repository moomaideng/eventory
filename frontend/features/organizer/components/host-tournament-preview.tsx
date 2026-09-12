"use client";

import React from "react";
import {
  CalendarDays,
  Clock3,
  MapPin,
  Users,
  WalletCards,
  Sparkles,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatEntryFee, formatTournamentDate } from "@/features/tournaments/utils";

export interface HostTournamentPreviewProps {
  name: string;
  game: string;
  location: string;
  description: string;
  registrationMode: "SOLO" | "TEAM";
  minTeamSize: number;
  maxTeamSize: number;
  capacity: number;
  entryFee: number;
  registrationDeadline: string;
  startAt: string;
  organizerName?: string;
}

export function HostTournamentPreview({
  name,
  game,
  location,
  description,
  registrationMode,
  minTeamSize,
  maxTeamSize,
  capacity,
  entryFee,
  registrationDeadline,
  startAt,
  organizerName,
}: HostTournamentPreviewProps) {
  const displayTitle = name.trim() || "Your Tournament Name";
  const displayGame = game.trim() || "Selected Game";
  const displayLocation = location.trim() || "Online";
  const displayDesc =
    description.trim() ||
    "Tournament description, rules, and participant guidelines will appear here.";

  const rosterLabel =
    registrationMode === "SOLO"
      ? "Solo (1v1)"
      : minTeamSize === maxTeamSize
        ? `Team (${minTeamSize} players)`
        : `Team (${minTeamSize}–${maxTeamSize} players)`;

  let formattedStart = "Schedule not set";
  if (startAt) {
    try {
      formattedStart = formatTournamentDate(new Date(startAt).toISOString());
    } catch {
      formattedStart = startAt;
    }
  }

  let formattedDeadline = "Deadline not set";
  if (registrationDeadline) {
    try {
      formattedDeadline = formatTournamentDate(
        new Date(registrationDeadline).toISOString()
      );
    } catch {
      formattedDeadline = registrationDeadline;
    }
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-2 text-xs font-semibold tracking-wider text-muted-foreground uppercase">
        <Sparkles className="size-3.5 text-primary" />
        Live Public Preview
      </div>

      <Card className="border-border/80 bg-card/60 relative flex flex-col justify-between backdrop-blur-xs transition-all">
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Badge variant="secondary">{displayGame}</Badge>
            <Badge variant="outline" className="capitalize">
              {rosterLabel}
            </Badge>
          </div>
          <CardTitle className="mt-2 text-xl font-bold wrap-anywhere">
            {displayTitle}
          </CardTitle>
          <CardDescription className="wrap-anywhere">
            Hosted by {organizerName || "Your Organization"}
          </CardDescription>
        </CardHeader>

        <CardContent className="flex flex-1 flex-col gap-4">
          <p className="text-muted-foreground line-clamp-3 text-sm wrap-anywhere">
            {displayDesc}
          </p>

          <div className="flex flex-col gap-2.5 text-sm">
            <div className="flex items-start gap-2">
              <CalendarDays className="text-muted-foreground mt-0.5 size-4 shrink-0" />
              <span>Starts: {formattedStart}</span>
            </div>
            <div className="flex items-start gap-2">
              <Clock3 className="text-muted-foreground mt-0.5 size-4 shrink-0" />
              <span>Closes: {formattedDeadline}</span>
            </div>
            <div className="flex items-start gap-2">
              <WalletCards className="text-muted-foreground mt-0.5 size-4 shrink-0" />
              <span>{formatEntryFee(entryFee, "THB")}</span>
            </div>
            <div className="flex items-start gap-2">
              <MapPin className="text-muted-foreground mt-0.5 size-4 shrink-0" />
              <span className="truncate">{displayLocation}</span>
            </div>
          </div>
        </CardContent>

        <CardFooter className="border-border/40 border-t pt-4">
          <div className="text-muted-foreground flex items-center gap-2 text-sm">
            <Users className="size-4" />
            <span>
              {capacity} {registrationMode === "SOLO" ? "players" : "teams"}{" "}
              capacity
            </span>
          </div>
        </CardFooter>
      </Card>
    </div>
  );
}
