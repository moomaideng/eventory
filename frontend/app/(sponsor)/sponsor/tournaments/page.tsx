import { Suspense } from "react";
import { TournamentCatalog } from "@/features/tournaments/components/tournament-catalog";
import { TournamentGridSkeleton } from "@/features/tournaments/components/tournament-grid-skeleton";

export const metadata = { title: "Find tournaments to sponsor | Eventory" };

export default function SponsorTournamentsPage() {
  return (
    <Suspense fallback={<TournamentGridSkeleton />}>
      <TournamentCatalog sponsorMode />
    </Suspense>
  );
}
