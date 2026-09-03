import { Suspense } from "react";
import { TournamentCatalog } from "./tournament-catalog";
import { Skeleton } from "@/components/ui/skeleton";

function CatalogFallback() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-col gap-8 px-4 py-12 sm:px-8">
      <div className="flex flex-col gap-3">
        <Skeleton className="h-10 w-72" />
        <Skeleton className="h-5 w-full max-w-xl" />
      </div>
      <Skeleton className="h-48 w-full" />
      <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Skeleton key={index} className="h-72 w-full" />
        ))}
      </div>
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
