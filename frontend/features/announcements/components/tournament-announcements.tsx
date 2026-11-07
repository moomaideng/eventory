"use client";

import React from "react";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { AnnouncementList } from "@/features/announcements/components/announcement-list";
import { useAnnouncements } from "@/features/announcements/hooks/use-announcements";
import { $api } from "@/lib/api/client";

export function TournamentAnnouncements({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const { data } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}",
    { params: { path: { tournamentId } } }
  );
  const tournament = data?.tournament;

  const { announcements, isLoading, isError, refetch } =
    useAnnouncements(tournamentId);

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-8 sm:px-8">
      <Button
        variant="outline"
        size="sm"
        render={<Link href={`/tournaments/${tournamentId}`} />}
        nativeButton={false}
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        Back to tournament
      </Button>

      <header className="flex flex-col gap-2">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline">Announcements</Badge>
          {tournament?.game ? (
            <Badge variant="secondary">{tournament.game}</Badge>
          ) : null}
        </div>
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          {tournament?.name
            ? `${tournament.name} Announcements`
            : "Tournament Announcements"}
        </h1>
        <p className="text-muted-foreground text-sm">
          Official updates and announcements from the tournament organizers.
        </p>
      </header>

      <main>
        <AnnouncementList
          announcements={announcements}
          isLoading={isLoading}
          isError={isError}
          onRetry={() => void refetch()}
          emptyDescription="No announcements have been published for this tournament yet."
        />
      </main>
    </div>
  );
}
