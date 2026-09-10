"use client";

import React from "react";
import {
  CalendarDays,
  Clock3,
  MapPin,
  Users,
  WalletCards,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  formatDateRange,
  formatMoney,
  formatTournamentDate,
  type Tournament,
} from "../utils";

interface TournamentAboutCardProps {
  tournament: Tournament;
  spotsLeft: number;
  registrationType: string;
}

export function TournamentAboutCard({
  tournament,
  spotsLeft,
  registrationType,
}: TournamentAboutCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>About this tournament</CardTitle>
        <CardDescription>
          What competitors and supporters should know.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-6">
        <p className="text-muted-foreground leading-7">
          {tournament.description}
        </p>
        <Separator />
        <div className="grid gap-5 sm:grid-cols-2">
          <div className="flex items-start gap-3">
            <CalendarDays className="text-muted-foreground mt-0.5 size-5 shrink-0" />
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium">Tournament schedule</p>
              <p className="text-muted-foreground text-sm">
                {formatDateRange(tournament.startAt, tournament.endAt)}
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <Clock3 className="text-muted-foreground mt-0.5 size-5 shrink-0" />
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium">Registration deadline</p>
              <p className="text-muted-foreground text-sm">
                {formatTournamentDate(tournament.registrationDeadline)}
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <MapPin className="text-muted-foreground mt-0.5 size-5 shrink-0" />
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium">Location</p>
              <p className="text-muted-foreground text-sm">
                {tournament.location}
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <WalletCards className="text-muted-foreground mt-0.5 size-5 shrink-0" />
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium">Entry fee</p>
              <p className="text-muted-foreground text-sm">
                {formatMoney(tournament.entryFee, tournament.currency, true)}
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <Users className="text-muted-foreground mt-0.5 size-5 shrink-0" />
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium">Availability</p>
              <p className="text-muted-foreground text-sm">
                {spotsLeft} of {tournament.capacity} spots available
              </p>
            </div>
          </div>
          <div className="flex items-start gap-3">
            <Users className="text-muted-foreground mt-0.5 size-5 shrink-0" />
            <div className="flex flex-col gap-1">
              <p className="text-sm font-medium">Registration type</p>
              <p className="text-muted-foreground text-sm">
                {registrationType}
              </p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
