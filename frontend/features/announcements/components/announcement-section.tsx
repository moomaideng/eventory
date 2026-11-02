"use client";

import { Info } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { AnnouncementForm } from "@/features/announcements/components/announcement-form";
import { AnnouncementList } from "@/features/announcements/components/announcement-list";
import { useAnnouncements } from "@/features/announcements/hooks/use-announcements";

export function AnnouncementSection({
  tournamentId,
  published,
}: {
  tournamentId: string;
  published: boolean;
}) {
  // A draft tournament has no public announcement list (the API answers 404),
  // so the request is skipped until the tournament is published.
  const { announcements, isLoading, isError, refetch } = useAnnouncements(
    tournamentId,
    published
  );

  return (
    <section
      className="flex min-w-0 flex-col gap-6"
      aria-labelledby="announcements-heading"
    >
      <div className="flex min-w-0 flex-col gap-1">
        <h2
          id="announcements-heading"
          className="text-xl font-bold tracking-tight"
        >
          Announcements
        </h2>
        <p className="text-muted-foreground text-sm">
          Keep players up to date about changes such as a new venue or schedule.
        </p>
      </div>

      {published ? (
        <>
          <AnnouncementForm tournamentId={tournamentId} />
          <AnnouncementList
            announcements={announcements}
            isLoading={isLoading}
            isError={isError}
            onRetry={() => void refetch()}
          />
        </>
      ) : (
        <Alert>
          <Info />
          <AlertTitle>Publish this tournament to post announcements</AlertTitle>
          <AlertDescription>
            Announcements appear on the public tournament page, so they become
            available once the tournament is published.
          </AlertDescription>
        </Alert>
      )}
    </section>
  );
}
