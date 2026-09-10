import React from "react";
import { Skeleton } from "@/components/ui/skeleton";

export function TournamentGridSkeleton() {
  return (
    <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: 6 }).map((_, index) => (
        <Skeleton key={index} className="h-72 w-full rounded-xl" />
      ))}
    </div>
  );
}
