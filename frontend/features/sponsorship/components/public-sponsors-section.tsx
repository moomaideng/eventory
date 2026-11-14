"use client";

import Image from "next/image";
import { ArrowUpRight, Handshake } from "lucide-react";
import { $api } from "@/lib/api/client";
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

export function PublicSponsorsSection({
  tournamentId,
  hideWhenEmpty = false,
}: {
  tournamentId: string;
  hideWhenEmpty?: boolean;
}) {
  const { data, isLoading } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/sponsors",
    {
      params: { path: { tournamentId } },
    },
    { staleTime: 15_000 }
  );

  if (isLoading) {
    return hideWhenEmpty ? null : <Skeleton className="h-44 rounded-xl" />;
  }

  const sponsors = data?.items ?? [];
  if (hideWhenEmpty && sponsors.length === 0) return null;

  return (
    <Card aria-labelledby="tournament-sponsors-heading">
      <CardHeader>
        <CardTitle id="tournament-sponsors-heading">
          Tournament sponsors
        </CardTitle>
        <CardDescription>
          {sponsors.length} sponsor{sponsors.length === 1 ? "" : "s"} currently
          backing this tournament.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {sponsors.length === 0 ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <Handshake />
              </EmptyMedia>
              <EmptyTitle>No sponsors yet</EmptyTitle>
              <EmptyDescription>
                Confirmed sponsors will appear here.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <ul className="grid gap-4 sm:grid-cols-2">
            {sponsors.map((sponsor) => (
              <li key={sponsor.id}>
                <a
                  href={sponsor.websiteUrl}
                  target="_blank"
                  rel="noopener noreferrer sponsored"
                  className="border-border bg-muted/40 hover:bg-muted/70 focus-visible:ring-ring flex items-center gap-4 rounded-xl border p-4 transition-colors outline-none focus-visible:ring-2"
                >
                  <Image
                    src={sponsor.logoUrl}
                    alt={`${sponsor.sponsorName} logo`}
                    width={96}
                    height={64}
                    unoptimized
                    className="bg-background h-16 w-24 shrink-0 rounded-md object-contain p-1"
                  />
                  <span className="flex min-w-0 flex-col gap-1">
                    <span className="font-semibold wrap-anywhere">
                      {sponsor.sponsorName}
                    </span>
                    <span className="text-muted-foreground text-sm">
                      {sponsor.tierName} sponsor
                    </span>
                    <span className="text-muted-foreground inline-flex items-center gap-1 text-xs">
                      Visit website{" "}
                      <ArrowUpRight className="size-3" aria-hidden="true" />
                    </span>
                  </span>
                </a>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
