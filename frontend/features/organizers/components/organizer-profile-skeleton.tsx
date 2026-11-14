import React from "react";
import { Skeleton } from "@/components/ui/skeleton";
import { TournamentGridSkeleton } from "@/features/tournaments/components/tournament-grid-skeleton";

export function OrganizerProfileSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <div className="flex flex-col gap-3">
        <Skeleton className="h-3 w-48" />
        <Skeleton className="h-9 w-full max-w-sm" />
        <Skeleton className="h-5 w-64" />
      </div>

      <Skeleton className="h-44 w-full rounded-xl" />

      <div className="flex flex-col gap-5">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <Skeleton className="h-7 w-48" />
          <Skeleton className="h-9 w-80" />
        </div>
        <TournamentGridSkeleton />
      </div>
    </div>
  );
}
