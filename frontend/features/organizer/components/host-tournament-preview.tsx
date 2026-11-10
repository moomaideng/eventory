"use client";

import React from "react";
import { MapPin, Sparkles, FileEdit } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { TournamentStatusBadge } from "@/features/tournaments/components/tournament-status-badge";

export interface HostTournamentPreviewProps {
  name: string;
  game: string;
  location: string;
  description: string;
  organizerName?: string;
}

export function HostTournamentPreview({
  name,
  game,
  location,
  description,
  organizerName,
}: HostTournamentPreviewProps) {
  const displayTitle = name.trim() || "Your Tournament Name";
  const displayGame = game.trim() || "Selected Game";
  const displayLocation = location.trim() || "Online";
  const displayDesc =
    description.trim() ||
    "Tournament description, rules, and participant guidelines will appear here.";

  return (
    <div className="flex flex-col gap-3">
      <div className="text-muted-foreground flex items-center gap-2 text-xs font-semibold tracking-wider uppercase">
        <Sparkles className="text-primary size-3.5" />
        Draft Tournament Preview
      </div>

      <Card className="border-border/80 bg-card/60 relative flex flex-col justify-between backdrop-blur-xs transition-all">
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-2">
            <Badge variant="secondary">{displayGame}</Badge>
            <TournamentStatusBadge status="DRAFT" />
          </div>
          <CardTitle className="mt-2 text-xl font-bold wrap-anywhere">
            {displayTitle}
          </CardTitle>
          <CardDescription className="wrap-anywhere">
            Hosted by {organizerName || "Your Organization"}
          </CardDescription>
        </CardHeader>

        <CardContent className="flex flex-1 flex-col gap-4">
          <p className="text-muted-foreground line-clamp-3 text-sm wrap-anywhere">
            {displayDesc}
          </p>

          <div className="flex flex-col gap-2.5 text-sm">
            <div className="flex items-start gap-2">
              <MapPin className="text-muted-foreground mt-0.5 size-4 shrink-0" />
              <span className="truncate">{displayLocation}</span>
            </div>
          </div>
        </CardContent>

        <CardFooter className="border-border/40 border-t pt-4">
          <div className="text-muted-foreground flex items-center gap-2 text-xs">
            <FileEdit className="size-3.5 shrink-0" />
            <span>
              Registration format, timeline, and funding rules can be customized
              in your workspace after creating this draft.
            </span>
          </div>
        </CardFooter>
      </Card>
    </div>
  );
}
