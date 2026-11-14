"use client";

import React from "react";
import Link from "next/link";
import { ChevronRight, Users } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import type { components } from "@/lib/api/schema";
import { initials } from "../utils";

type Team = components["schemas"]["TournamentTeamResponse"];

export function TournamentRegisteredTeamsCard({
  tournamentId,
  teams,
  isSolo,
}: {
  tournamentId: string;
  teams: Team[];
  isSolo: boolean;
}) {
  const itemLabel = isSolo ? "competitor" : "team";

  return (
    <Card>
      <CardHeader>
        <CardTitle>Registered {isSolo ? "competitors" : "teams"}</CardTitle>
        <CardDescription>
          {teams.length} {itemLabel}
          {teams.length === 1 ? "" : "s"} currently listed for this tournament.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {teams.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Users />
              </EmptyMedia>
              <EmptyTitle>No {itemLabel}s listed yet</EmptyTitle>
              <EmptyDescription>
                Confirmed {itemLabel}s will appear here.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {teams.map((team) => {
              const content = (
                <>
                  <Avatar size="lg">
                    <AvatarFallback>{initials(team.name)}</AvatarFallback>
                  </Avatar>
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">{team.name}</p>
                    <p className="text-muted-foreground text-sm">
                      {isSolo
                        ? "Confirmed competitor"
                        : `${team.memberCount} member${team.memberCount === 1 ? "" : "s"}`}
                    </p>
                  </div>
                  {!isSolo ? (
                    <ChevronRight className="text-muted-foreground size-4 shrink-0" />
                  ) : null}
                </>
              );

              return isSolo ? (
                <div
                  key={team.id}
                  className="bg-muted/40 flex items-center gap-3 rounded-lg p-3"
                >
                  {content}
                </div>
              ) : (
                <Link
                  key={team.id}
                  href={`/tournaments/${tournamentId}/teams/${team.id}`}
                  className="bg-muted/40 hover:bg-muted/70 flex items-center gap-3 rounded-lg p-3 transition-colors"
                >
                  {content}
                </Link>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
