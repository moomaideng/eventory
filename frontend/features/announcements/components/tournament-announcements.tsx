"use client";

import React from "react";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { AnnouncementList } from "@/features/announcements/components/announcement-list";
import { useAnnouncements } from "@/features/announcements/hooks/use-announcements";
import { $api } from "@/lib/api/client";

export function TournamentAnnouncements({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const { data } = $api.useQuery("get", "/api/v1/tournaments/{tournamentId}", {
    params: { path: { tournamentId } },
  });
  const tournament = data?.tournament;

  const { announcements, isLoading, isError, refetch } =
    useAnnouncements(tournamentId);

  React.useEffect(() => {
    // The feed loads asynchronously, after the browser's initial hash scroll.
    const showSelected = () => {
      const selected = announcements.find(
        (item) => window.location.hash === `#announcement-${item.id}`
      );
      if (!selected) return;
      const element = document.getElementById(`announcement-${selected.id}`);
      element?.scrollIntoView({ block: "center" });
      element?.focus({ preventScroll: true });
    };
    showSelected();
    window.addEventListener("hashchange", showSelected);
    return () => window.removeEventListener("hashchange", showSelected);
  }, [announcements]);

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-6 px-4 py-8 sm:px-8">
      <PageHeader
        breadcrumbs={[
          { label: "Tournaments", href: "/tournaments" },
          {
            label: tournament?.name ?? "Tournament",
            href: `/tournaments/${tournamentId}`,
          },
          { label: "Announcements" },
        ]}
        tags={
          <>
            <Badge variant="outline">Announcements</Badge>
            {tournament?.game ? (
              <Badge variant="secondary">{tournament.game}</Badge>
            ) : null}
          </>
        }
        title={
          tournament?.name
            ? `${tournament.name} announcements`
            : "Tournament announcements"
        }
        description="Official updates and announcements from the tournament organizers."
      />

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
