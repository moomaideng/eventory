import { AlertCircle, Megaphone } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
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
import { Skeleton } from "@/components/ui/skeleton";
import type { Announcement } from "@/features/announcements/hooks/use-announcements";
import { formatTournamentDate } from "@/features/tournaments/utils";

interface AnnouncementListProps {
  announcements: Announcement[];
  isLoading: boolean;
  isError: boolean;
  onRetry: () => void;
  emptyDescription?: string;
}

export function AnnouncementList({
  announcements,
  isLoading,
  isError,
  onRetry,
  emptyDescription,
}: AnnouncementListProps) {
  if (isLoading) {
    return (
      <div className="flex flex-col gap-3">
        <Skeleton className="h-24 w-full rounded-xl" />
        <Skeleton className="h-24 w-full rounded-xl" />
      </div>
    );
  }

  if (isError) {
    return (
      <Alert variant="destructive">
        <AlertCircle />
        <AlertTitle>Unable to load announcements</AlertTitle>
        <AlertDescription className="flex items-center justify-between gap-4">
          <span>Something went wrong while loading the announcements.</span>
          <Button variant="outline" size="sm" onClick={onRetry}>
            Retry
          </Button>
        </AlertDescription>
      </Alert>
    );
  }

  if (announcements.length === 0) {
    return (
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Megaphone />
          </EmptyMedia>
          <EmptyTitle>No announcements yet</EmptyTitle>
          <EmptyDescription>
            {emptyDescription ??
              "Announcements you post will appear here, newest first."}
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <ul className="flex flex-col gap-3">
      {announcements.map((announcement) => (
        <li key={announcement.id} id={`announcement-${announcement.id}`}>
          <Card>
            <CardHeader>
              <CardTitle>{announcement.title}</CardTitle>
              <CardDescription>
                {announcement.authorName} ·{" "}
                {formatTournamentDate(announcement.createdAt)}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-sm whitespace-pre-wrap">{announcement.body}</p>
            </CardContent>
          </Card>
        </li>
      ))}
    </ul>
  );
}
