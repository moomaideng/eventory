"use client";

import React from "react";
import { CalendarDays, MapPin } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { formatDate, type Tournament } from "../utils";

export function TournamentSummaryCard({
  tournament,
}: {
  tournament: Tournament;
}) {
  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex flex-col gap-1">
            <CardTitle>{tournament.name}</CardTitle>
            <CardDescription>
              {tournament.game} · Hosted by {tournament.organizerName}
            </CardDescription>
          </div>
          <Badge variant="outline">
            {tournament.minTeamSize}–{tournament.maxTeamSize} players
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-3 text-sm">
        <p className="text-muted-foreground">{tournament.description}</p>
        <div className="text-muted-foreground flex flex-wrap gap-x-5 gap-y-2">
          <span className="flex items-center gap-2">
            <CalendarDays className="size-4 shrink-0" />
            Registration closes {formatDate(tournament.registrationDeadline)}
          </span>
          <span className="flex items-center gap-2">
            <MapPin className="size-4 shrink-0" />
            {tournament.location}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
