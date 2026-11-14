"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";
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
import { AnnouncementList } from "@/features/announcements/components/announcement-list";
import { useAnnouncements } from "@/features/announcements/hooks/use-announcements";
import { formatTournamentDate } from "@/features/tournaments/utils";

export function AnnouncementPreview({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const { announcements, isLoading, isError, refetch } =
    useAnnouncements(tournamentId);
  const href = `/tournaments/${tournamentId}/announcements`;

  return (
    <Card aria-label="Tournament announcements">
      <CardHeader>
        <CardTitle className="flex items-center justify-between gap-2">
          <span>Announcements</span>
          {!isLoading && !isError ? (
            <Badge variant="secondary">{announcements.length}</Badge>
          ) : null}
        </CardTitle>
        <CardDescription>
          Latest updates from the tournament organizer.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {isLoading || isError || announcements.length === 0 ? (
          <AnnouncementList
            announcements={[]}
            isLoading={isLoading}
            isError={isError}
            onRetry={() => void refetch()}
            emptyDescription="Updates from the organizer will appear here."
          />
        ) : (
          <ul className="flex flex-col gap-2">
            {announcements.slice(0, 3).map((announcement) => (
              <li key={announcement.id}>
                <Link
                  href={`${href}#announcement-${announcement.id}`}
                  className="hover:bg-muted focus-visible:ring-ring flex items-start gap-3 rounded-lg border p-3 focus-visible:ring-2 focus-visible:outline-none"
                >
                  <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <span className="line-clamp-2 font-medium wrap-anywhere">
                      {announcement.title}
                    </span>
                    <span className="text-muted-foreground text-xs">
                      {formatTournamentDate(announcement.createdAt)}
                    </span>
                    <span className="text-muted-foreground line-clamp-2 text-sm wrap-anywhere">
                      {announcement.body}
                    </span>
                  </div>
                  <ArrowRight
                    className="mt-1 size-4 shrink-0"
                    aria-hidden="true"
                  />
                </Link>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
      <CardFooter>
        <Button
          variant="outline"
          className="w-full"
          render={<Link href={href} />}
          nativeButton={false}
        >
          View more
          <ArrowRight data-icon="inline-end" />
        </Button>
      </CardFooter>
    </Card>
  );
}
