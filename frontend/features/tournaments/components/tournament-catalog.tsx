"use client";

import React from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { ArrowLeft, ArrowRight, Trophy } from "lucide-react";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { PAGE_SIZE } from "../utils";
import { TournamentCard } from "./tournament-card";
import { TournamentFiltersCard } from "./tournament-filters-card";
import { TournamentGridSkeleton } from "./tournament-grid-skeleton";

export function TournamentCatalog() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const requestedPage = Number(searchParams.get("page") ?? "1");
  const page =
    Number.isInteger(requestedPage) && requestedPage > 0 ? requestedPage : 1;
  const rawMaxFee = searchParams.get("maxEntryFee");
  const parsedMaxFee = rawMaxFee ? Number(rawMaxFee) : undefined;
  const maxEntryFee =
    typeof parsedMaxFee === "number" &&
    !Number.isNaN(parsedMaxFee) &&
    parsedMaxFee >= 0
      ? parsedMaxFee
      : undefined;

  const { data, error, isLoading, isFetching } = $api.useQuery(
    "get",
    "/api/v1/tournaments",
    {
      params: {
        query: {
          q: searchParams.get("q") || undefined,
          startFrom: searchParams.get("startFrom") || undefined,
          startTo: searchParams.get("startTo") || undefined,
          maxEntryFee,
          sort: "start_asc",
          page,
          pageSize: PAGE_SIZE,
        },
      },
    },
    { staleTime: 30_000 }
  );

  const items = data?.items ?? [];
  const pageCount = data
    ? Math.max(1, Math.ceil(data.total / data.pageSize))
    : 1;

  function applyFilters(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const formData = new FormData(event.currentTarget);
    const query = String(formData.get("q") ?? "").trim();
    const startFrom = String(formData.get("startFrom") ?? "");
    const startTo = String(formData.get("startTo") ?? "");
    const maxEntryFee = String(formData.get("maxEntryFee") ?? "");
    if (startFrom && startTo && startFrom > startTo) {
      return;
    }
    const next = new URLSearchParams();
    if (query) next.set("q", query);
    if (startFrom) next.set("startFrom", startFrom);
    if (startTo) next.set("startTo", startTo);
    if (maxEntryFee) next.set("maxEntryFee", maxEntryFee);
    router.push(next.size ? `${pathname}?${next.toString()}` : pathname);
  }

  function clearFilters() {
    router.push(pathname);
  }

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams.toString());
    if (nextPage <= 1) {
      next.delete("page");
    } else {
      next.set("page", String(nextPage));
    }
    router.push(next.size ? `${pathname}?${next.toString()}` : pathname);
  }

  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-12 sm:px-8">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
          Find your next tournament
        </h1>
        <p className="text-muted-foreground max-w-2xl">
          Search published competitions and narrow the list to events that fit
          your schedule and budget.
        </p>
      </div>

      <TournamentFiltersCard
        key={searchParams.toString()}
        searchKey={searchParams.toString()}
        defaultQ={searchParams.get("q") ?? ""}
        defaultStartFrom={searchParams.get("startFrom") ?? ""}
        defaultStartTo={searchParams.get("startTo") ?? ""}
        defaultMaxEntryFee={searchParams.get("maxEntryFee") ?? ""}
        onApplyFilters={applyFilters}
        onClearFilters={clearFilters}
      />

      <div className="flex items-center justify-between gap-4">
        <p className="text-sm font-medium">
          {data
            ? `${data.total} tournament${data.total === 1 ? "" : "s"}`
            : "Loading tournaments"}
        </p>
        {isFetching && !isLoading ? (
          <span className="text-muted-foreground text-sm">Updating…</span>
        ) : null}
      </div>

      {error ? (
        <Alert variant="destructive">
          <AlertTitle>Could not load tournaments</AlertTitle>
          <AlertDescription>
            Check your filters or try again after the API is available.
          </AlertDescription>
        </Alert>
      ) : isLoading ? (
        <TournamentGridSkeleton />
      ) : items.length === 0 ? (
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Trophy />
            </EmptyMedia>
            <EmptyTitle>No tournaments match</EmptyTitle>
            <EmptyDescription>
              Try widening the date range, increasing the budget, or using a
              different search term.
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent>
            <Button variant="outline" onClick={clearFilters}>
              Clear filters
            </Button>
          </EmptyContent>
        </Empty>
      ) : (
        <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
          {items.map((tournament) => (
            <TournamentCard key={tournament.id} tournament={tournament} />
          ))}
        </div>
      )}

      {data && data.total > data.pageSize ? (
        <div className="flex items-center justify-center gap-3">
          <Button
            variant="outline"
            disabled={page <= 1 || isFetching}
            onClick={() => goToPage(page - 1)}
          >
            <ArrowLeft data-icon="inline-start" />
            Previous
          </Button>
          <span className="text-muted-foreground text-sm">
            Page {page} of {pageCount}
          </span>
          <Button
            variant="outline"
            disabled={page >= pageCount || isFetching}
            onClick={() => goToPage(page + 1)}
          >
            Next
            <ArrowRight data-icon="inline-end" />
          </Button>
        </div>
      ) : null}
    </div>
  );
}
