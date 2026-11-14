"use client";

import { RefreshCw } from "lucide-react";
import { $api } from "@/lib/api/client";
import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { TournamentHeaderTags } from "@/features/tournaments/components/tournament-header-tags";
import { BracketTreeView } from "./bracket-tree-view";
import { useBracket } from "../hooks/use-bracket";

interface PublicBracketViewerProps {
  tournamentId: string;
}

export function PublicBracketViewer({ tournamentId }: PublicBracketViewerProps) {
  const { matches, isLoading, isFetching, refetch } = useBracket({ tournamentId });
  const { data: tournamentData } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}",
    { params: { path: { tournamentId } } },
    { staleTime: 30_000 }
  );
  const tournament = tournamentData?.tournament;

  return (
    <div className="flex min-w-0 flex-col gap-6">
      <div className="mx-auto w-full max-w-7xl px-4 sm:px-8">
        <PageHeader
          breadcrumbs={[
            { label: "Tournaments", href: "/tournaments" },
            {
              label: tournament?.name ?? "Tournament",
              href: `/tournaments/${tournamentId}`,
            },
            { label: "Bracket" },
          ]}
          tags={
            tournament ? (
              <TournamentHeaderTags
                game={tournament.game}
                status={tournament.status}
                registrationMode={tournament.registrationMode}
                registrationDeadline={tournament.registrationDeadline}
              />
            ) : undefined
          }
          title={
            tournament?.name
              ? `${tournament.name} bracket`
              : "Tournament bracket"
          }
          actions={
            <Button
              variant="outline"
              size="sm"
              onClick={() => void refetch()}
              disabled={isFetching}
            >
              <RefreshCw
                data-icon="inline-start"
                className={isFetching ? "animate-spin" : undefined}
              />
              Refresh
            </Button>
          }
        />
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3">
          <div className="size-7 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="text-xs">Loading tournament bracket...</p>
        </div>
      ) : (
        <div className="w-full border-y border-border/80 bg-card px-4 py-6 shadow-xs sm:px-8">
          <BracketTreeView matches={matches} canEdit={false} />
        </div>
      )}
    </div>
  );
}
