import React from "react";
import { Skeleton } from "@/components/ui/skeleton";

export function TournamentDetailsSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <div className="flex flex-col gap-3">
        <Skeleton className="h-3 w-48" />
        <Skeleton className="h-5 w-36" />
        <Skeleton className="h-9 w-full max-w-xl" />
        <Skeleton className="h-5 w-56" />
      </div>
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex flex-col gap-6">
          <Skeleton className="h-52 w-full rounded-xl" />
          <Skeleton className="h-80 w-full rounded-xl" />
        </div>
        <Skeleton className="h-96 w-full rounded-xl" />
      </div>
    </div>
  );
}
