"use client";

import Link from "next/link";
import { ArrowLeft, RefreshCw, Trophy } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { BracketTreeView } from "./bracket-tree-view";
import { useBracket } from "../hooks/use-bracket";

interface PublicBracketViewerProps {
  tournamentId: string;
}

export function PublicBracketViewer({ tournamentId }: PublicBracketViewerProps) {
  const { matches, isLoading, isFetching, refetch } = useBracket({ tournamentId });

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between flex-wrap gap-4">
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            render={<Link href={`/tournaments/${tournamentId}`} />}
            nativeButton={false}
          >
            <ArrowLeft data-icon="inline-start" />
            Tournament Details
          </Button>

          <div className="flex items-center gap-2">
            <Trophy className="size-5 text-primary" />
            <h1 className="text-xl font-bold">Tournament Bracket</h1>
            {matches.length > 0 ? (
              <Badge variant="outline" className="font-mono text-xs">
                {matches.length} {matches.length === 1 ? "match" : "matches"}
              </Badge>
            ) : null}
          </div>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={() => void refetch()}
          disabled={isFetching}
          className="gap-2 text-xs"
        >
          <RefreshCw className={`size-3.5 ${isFetching ? "animate-spin" : ""}`} />
          Refresh
        </Button>
      </div>

      {isLoading ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3">
          <div className="size-7 animate-spin rounded-full border-2 border-primary border-t-transparent" />
          <p className="text-xs">Loading tournament bracket...</p>
        </div>
      ) : (
        <div className="rounded-xl border border-border/80 bg-card p-6 shadow-xs">
          <BracketTreeView matches={matches} canEdit={false} />
        </div>
      )}
    </div>
  );
}
