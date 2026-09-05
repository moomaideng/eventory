"use client";

import React from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import {
  ArrowLeft,
  ArrowRight,
  CalendarDays,
  Clock3,
  MapPin,
  Search,
  Trophy,
  Users,
  WalletCards,
  ArrowUpRight,
} from "lucide-react";
import { $api } from "@/lib/api/client";
import type { components } from "@/lib/api/schema";
import { useRole } from "@/context/role-context";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";

const PAGE_SIZE = 6;

export function TournamentCatalog() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { activeRole } = useRole();

  const requestedPage = Number(searchParams.get("page") ?? "1");
  const page =
    Number.isInteger(requestedPage) && requestedPage > 0 ? requestedPage : 1;
  const maxFeeValue = searchParams.get("maxEntryFee");

  const { data, error, isLoading, isFetching } = $api.useQuery(
    "get",
    "/api/v1/tournaments",
    {
      params: {
        query: {
          q: searchParams.get("q") || undefined,
          startFrom: searchParams.get("startFrom") || undefined,
          startTo: searchParams.get("startTo") || undefined,
          maxEntryFee: maxFeeValue ? Number(maxFeeValue) : undefined,
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
      <div className="flex flex-col gap-3">
        <Badge variant="outline" className="capitalize">
          {activeRole} catalog
        </Badge>
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
            Find your next tournament
          </h1>
          <p className="text-muted-foreground max-w-2xl">
            Search published competitions and narrow the list to events that fit
            your schedule and budget.
          </p>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Search and filters</CardTitle>
          <CardDescription>
            Dates and times use Bangkok time. Entry fees are currently listed in
            Thai baht.
          </CardDescription>
        </CardHeader>
        <form key={searchParams.toString()} onSubmit={applyFilters}>
          <CardContent>
            <FieldGroup className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
              <Field>
                <FieldLabel htmlFor="tournament-search">Search</FieldLabel>
                <Input
                  id="tournament-search"
                  name="q"
                  defaultValue={searchParams.get("q") ?? ""}
                  placeholder="Name, game, or keyword"
                  maxLength={100}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="start-from">Starts after</FieldLabel>
                <Input
                  id="start-from"
                  name="startFrom"
                  type="date"
                  defaultValue={searchParams.get("startFrom") ?? ""}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="start-to">Starts before</FieldLabel>
                <Input
                  id="start-to"
                  name="startTo"
                  type="date"
                  defaultValue={searchParams.get("startTo") ?? ""}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="maximum-fee">Maximum fee (THB)</FieldLabel>
                <Input
                  id="maximum-fee"
                  name="maxEntryFee"
                  type="number"
                  min={0}
                  step={1}
                  defaultValue={searchParams.get("maxEntryFee") ?? ""}
                  placeholder="Any budget"
                />
              </Field>
            </FieldGroup>
          </CardContent>
          <CardFooter className="flex flex-wrap justify-end gap-2">
            <Button type="button" variant="ghost" onClick={clearFilters}>
              Clear filters
            </Button>
            <Button type="submit">
              <Search data-icon="inline-start" />
              Apply filters
            </Button>
          </CardFooter>
        </form>
      </Card>

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

type Tournament = components["schemas"]["TournamentResponse"];

function TournamentCard({ tournament }: { tournament: Tournament }) {
  const seatsLeft = Math.max(
    0,
    tournament.capacity - tournament.registeredCount
  );
  const status = tournament.status.replaceAll("_", " ").toLowerCase();

  return (
    <Card className="h-full">
      <CardHeader>
        <div className="flex items-start justify-between gap-3">
          <Badge variant="secondary">{tournament.game}</Badge>
          <Badge variant="outline" className="capitalize">
            {status}
          </Badge>
        </div>
        <CardTitle>{tournament.name}</CardTitle>
        <CardDescription>Hosted by {tournament.organizerName}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-4">
        <p className="text-muted-foreground line-clamp-3 text-sm">
          {tournament.description}
        </p>
        <div className="flex flex-col gap-2 text-sm">
          <Detail icon={CalendarDays}>
            {formatTournamentDate(tournament.startAt)}
          </Detail>
          <Detail icon={Clock3}>
            Registration closes{" "}
            {formatTournamentDate(tournament.registrationDeadline)}
          </Detail>
          <Detail icon={WalletCards}>
            {formatEntryFee(tournament.entryFee, tournament.currency)}
          </Detail>
          <Detail icon={MapPin}>{tournament.location}</Detail>
        </div>
      </CardContent>
      <CardFooter className="flex justify-between gap-3">
        <div className="text-muted-foreground flex items-center gap-2 text-sm">
          <Users className="size-4" />
          {seatsLeft} of {tournament.capacity} spots available
        </div>
        <Button
          render={<Link href={`/tournaments/${tournament.id}`} />}
          nativeButton={false}
          size="sm"
          variant="outline"
        >
          View details
          <ArrowUpRight data-icon="inline-end" />
        </Button>
      </CardFooter>
    </Card>
  );
}

function Detail({
  icon: Icon,
  children,
}: {
  icon: React.ElementType;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-2">
      <Icon className="text-muted-foreground mt-0.5 size-4 shrink-0" />
      <span>{children}</span>
    </div>
  );
}

function TournamentGridSkeleton() {
  return (
    <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: 3 }).map((_, index) => (
        <Skeleton key={index} className="h-72 w-full" />
      ))}
    </div>
  );
}

function formatTournamentDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Bangkok",
  }).format(new Date(value));
}

function formatEntryFee(entryFee: number, currency: string) {
  if (entryFee === 0) return "Free entry";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(entryFee);
}
