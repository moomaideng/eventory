"use client";

import React from "react";
import Link from "next/link";
import {
  CalendarDays,
  Clock3,
  MapPin,
  Users,
  WalletCards,
  ArrowUpRight,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  formatEntryFee,
  formatTournamentDate,
  type Tournament,
} from "../utils";

export function TournamentCard({ tournament }: { tournament: Tournament }) {
  const seatsLeft = Math.max(
    0,
    tournament.capacity - tournament.registeredCount
  );
  const status = tournament.status.replaceAll("_", " ").toLowerCase();

  return (
    <Card className="flex h-full flex-col justify-between">
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <Badge variant="secondary">{tournament.game}</Badge>
          <Badge variant="outline" className="capitalize">
            {status}
          </Badge>
        </div>
        <CardTitle>{tournament.name}</CardTitle>
        <CardDescription>Hosted by {tournament.organizerName}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-4">
        <p className="text-muted-foreground line-clamp-3 text-sm">
          {tournament.description}
        </p>
        <div className="flex flex-col gap-2 text-sm">
          <div className="flex items-start gap-2">
            <CalendarDays className="text-muted-foreground mt-0.5 size-4 shrink-0" />
            <span>{formatTournamentDate(tournament.startAt)}</span>
          </div>
          <div className="flex items-start gap-2">
            <Clock3 className="text-muted-foreground mt-0.5 size-4 shrink-0" />
            <span>
              Registration closes{" "}
              {formatTournamentDate(tournament.registrationDeadline)}
            </span>
          </div>
          <div className="flex items-start gap-2">
            <WalletCards className="text-muted-foreground mt-0.5 size-4 shrink-0" />
            <span>{formatEntryFee(tournament.entryFee, tournament.currency)}</span>
          </div>
          <div className="flex items-start gap-2">
            <MapPin className="text-muted-foreground mt-0.5 size-4 shrink-0" />
            <span>{tournament.location}</span>
          </div>
        </div>
      </CardContent>
      <CardFooter className="flex justify-between gap-3">
        <div className="text-muted-foreground flex items-center gap-2 text-sm">
          <Users className="size-4" />
          <span>
            {seatsLeft} of {tournament.capacity} spots available
          </span>
        </div>
        <Button
          render={<Link href={`/tournaments/${tournament.id}`} />}
          nativeButton={false}
          size="sm"
          variant="outline"
        >
          View details
          <ArrowUpRight data-icon="inline-end" />
        </Button>
      </CardFooter>
    </Card>
  );
}
