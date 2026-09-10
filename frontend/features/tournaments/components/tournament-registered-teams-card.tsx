"use client";

import React from "react";
import { Users } from "lucide-react";
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

export function TournamentRegisteredTeamsCard({ teams }: { teams: Team[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Registered teams</CardTitle>
        <CardDescription>
          {teams.length} team{teams.length === 1 ? "" : "s"} currently listed for
          this tournament.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {teams.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Users />
              </EmptyMedia>
              <EmptyTitle>No teams listed yet</EmptyTitle>
              <EmptyDescription>
                Registered teams will appear here when they are confirmed.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {teams.map((team) => (
              <div
                key={team.id}
                className="bg-muted/40 flex items-center gap-3 rounded-lg p-3"
              >
                <Avatar size="lg">
                  <AvatarFallback>{initials(team.name)}</AvatarFallback>
                </Avatar>
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{team.name}</p>
                  <p className="text-muted-foreground text-sm">
                    {team.memberCount} member
                    {team.memberCount === 1 ? "" : "s"}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
