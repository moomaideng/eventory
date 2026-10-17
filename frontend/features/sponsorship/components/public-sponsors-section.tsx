"use client";

import Image from "next/image";
import { ArrowUpRight } from "lucide-react";
import { $api, apiBaseUrl } from "@/lib/api/client";

export function PublicSponsorsSection({
  tournamentId,
}: {
  tournamentId: string;
}) {
  const { data } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}/sponsors",
    {
      params: { path: { tournamentId } },
    },
    { staleTime: 15_000 }
  );
  if (!data?.items?.length) return null;
  return (
    <section
      className="flex flex-col gap-4"
      aria-labelledby="tournament-sponsors-heading"
    >
      <h2
        id="tournament-sponsors-heading"
        className="text-xl font-bold tracking-tight"
      >
        Tournament sponsors
      </h2>
      <ul className="grid gap-4 sm:grid-cols-2">
        {data.items.map((sponsor) => (
          <li key={sponsor.id}>
            <a
              href={sponsor.websiteUrl}
              target="_blank"
              rel="noopener noreferrer sponsored"
              className="border-border hover:bg-muted focus-visible:ring-ring flex items-center gap-4 rounded-xl border p-4 outline-none focus-visible:ring-2"
            >
              <Image
                src={`${apiBaseUrl || "http://localhost:8080"}${sponsor.logoPath}`}
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
    </section>
  );
}
