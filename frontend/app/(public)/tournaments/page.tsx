import React, { Suspense } from "react";
import type { Metadata } from "next";
import { TournamentCatalog } from "@/features/tournaments/components/tournament-catalog";
import { TournamentGridSkeleton } from "@/features/tournaments/components/tournament-grid-skeleton";
import { Skeleton } from "@/components/ui/skeleton";

export const metadata: Metadata = {
  title: "Tournaments - Eventory",
  description:
    "Browse, search, and filter open esports and gaming tournaments.",
};

function CatalogFallback() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 py-12 sm:px-8">
      <div className="flex flex-col gap-3">
        <Skeleton className="h-10 w-72" />
        <Skeleton className="h-5 w-full max-w-xl" />
      </div>
      <Skeleton className="h-48 w-full" />
      <TournamentGridSkeleton />
    </div>
  );
}

export default function TournamentsPage() {
  return (
    <Suspense fallback={<CatalogFallback />}>
      <TournamentCatalog />
    </Suspense>
  );
}
