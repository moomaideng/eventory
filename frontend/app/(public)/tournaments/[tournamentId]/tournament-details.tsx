"use client";

import Link from "next/link";
import {
  ArrowLeft,
  CalendarDays,
  Clock3,
  MapPin,
  Trophy,
  Users,
  WalletCards,
} from "lucide-react";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
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
import {
  Progress,
  ProgressLabel,
  ProgressValue,
} from "@/components/ui/progress";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";

export function TournamentDetails({ tournamentId }: { tournamentId: string }) {
  const { data, error, isLoading } = $api.useQuery(
    "get",
    "/api/v1/tournaments/{tournamentId}",
    { params: { path: { tournamentId } } },
    { staleTime: 30_000 }
  );

  if (isLoading) return <DetailsSkeleton />;

  if (error || !data) {
    return (
      <div className="container mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-12 sm:px-8">
        <Alert variant="destructive">
          <AlertTitle>Tournament unavailable</AlertTitle>
          <AlertDescription>
            This tournament may not exist, may be unpublished, or the API may be
            temporarily unavailable.
          </AlertDescription>
        </Alert>
        <Button
          render={<Link href="/tournaments" />}
          nativeButton={false}
          variant="outline"
          className="self-start"
        >
          <ArrowLeft data-icon="inline-start" />
          Back to tournaments
        </Button>
      </div>
    );
  }

  const { tournament, funding } = data;
  const teams = data.teams ?? [];
  const spotsLeft = Math.max(
    tournament.capacity - tournament.registeredCount,
    0
  );
  const status = tournament.status.replaceAll("_", " ").toLowerCase();
  const progressValue = Math.min(Math.max(funding.percentage, 0), 100);

  return (
    <main className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <Button
        render={<Link href="/tournaments" />}
        nativeButton={false}
        variant="ghost"
        className="self-start"
      >
        <ArrowLeft data-icon="inline-start" />
        All tournaments
      </Button>

      <header className="flex flex-col gap-4">
        <div className="flex flex-wrap gap-2">
          <Badge>{tournament.game}</Badge>
          <Badge variant="outline" className="capitalize">
            {status}
          </Badge>
        </div>
        <div className="flex flex-col gap-2">
          <h1 className="max-w-4xl text-3xl font-bold tracking-tight sm:text-5xl">
            {tournament.name}
          </h1>
          <p className="text-muted-foreground text-base">
            Hosted by {tournament.organizerName}
          </p>
        </div>
      </header>

      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex flex-col gap-6">
          <Card>
            <CardHeader>
              <CardTitle>About this tournament</CardTitle>
              <CardDescription>
                What competitors and supporters should know.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground leading-7">
                {tournament.description}
              </p>
              <Separator />
              <div className="grid gap-5 sm:grid-cols-2">
                <Detail icon={CalendarDays} label="Tournament schedule">
                  {formatDateRange(tournament.startAt, tournament.endAt)}
                </Detail>
                <Detail icon={Clock3} label="Registration deadline">
                  {formatTournamentDate(tournament.registrationDeadline)}
                </Detail>
                <Detail icon={MapPin} label="Location">
                  {tournament.location}
                </Detail>
                <Detail icon={WalletCards} label="Entry fee">
                  {formatMoney(tournament.entryFee, tournament.currency, true)}
                </Detail>
                <Detail icon={Users} label="Availability">
                  {spotsLeft} of {tournament.capacity} spots available
                </Detail>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Registered teams</CardTitle>
              <CardDescription>
                {teams.length} team{teams.length === 1 ? "" : "s"} currently
                listed for this tournament.
              </CardDescription>
            </CardHeader>
            <CardContent>
              {teams.length === 0 ? (
                <Empty>
                  <EmptyHeader>
                    <EmptyMedia variant="icon">
                      <Users />
                    </EmptyMedia>
                    <EmptyTitle>No teams listed yet</EmptyTitle>
                    <EmptyDescription>
                      Registered teams will appear here when they are confirmed.
                    </EmptyDescription>
                  </EmptyHeader>
                </Empty>
              ) : (
                <div className="grid gap-3 sm:grid-cols-2">
                  {teams.map((team) => (
                    <div
                      key={team.id}
                      className="bg-muted/40 flex items-center gap-3 rounded-lg p-3"
                    >
                      <Avatar size="lg">
                        <AvatarFallback>{initials(team.name)}</AvatarFallback>
                      </Avatar>
                      <div className="min-w-0 flex-1">
                        <p className="truncate font-medium">{team.name}</p>
                        <p className="text-muted-foreground text-sm">
                          {team.memberCount} member
                          {team.memberCount === 1 ? "" : "s"}
                        </p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <Card className="lg:sticky lg:top-24">
          <CardHeader>
            <div className="bg-primary text-primary-foreground flex size-10 items-center justify-center rounded-lg">
              <Trophy className="size-5" />
            </div>
            <CardTitle>Funding progress</CardTitle>
            <CardDescription>
              Community support helps fund the prize pool and event costs.
            </CardDescription>
          </CardHeader>
          <CardContent className="gap-5">
            <div className="flex flex-col gap-1">
              <p className="text-3xl font-bold tracking-tight">
                {formatMoney(funding.raisedAmount, funding.currency)}
              </p>
              <p className="text-muted-foreground text-sm">
                raised of {formatMoney(funding.goalAmount, funding.currency)}
              </p>
            </div>
            <Progress value={progressValue}>
              <ProgressLabel>Funding goal</ProgressLabel>
              <ProgressValue>
                {() => formatPercentage(funding.percentage)}
              </ProgressValue>
            </Progress>
            <Separator />
            <div className="grid grid-cols-2 gap-4">
              <FundingStat
                label="Supporters"
                value={new Intl.NumberFormat("en-US").format(
                  funding.supporterCount
                )}
              />
              <FundingStat
                label="Still needed"
                value={formatMoney(funding.remainingAmount, funding.currency)}
              />
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  );
}

function Detail({
  icon: Icon,
  label,
  children,
}: {
  icon: React.ElementType;
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3">
      <Icon className="text-muted-foreground mt-0.5 size-5 shrink-0" />
      <div className="flex flex-col gap-1">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-muted-foreground text-sm">{children}</p>
      </div>
    </div>
  );
}

function FundingStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex flex-col gap-1">
      <p className="text-muted-foreground text-xs">{label}</p>
      <p className="font-medium">{value}</p>
    </div>
  );
}

function DetailsSkeleton() {
  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <Skeleton className="h-9 w-36" />
      <div className="flex flex-col gap-4">
        <Skeleton className="h-6 w-28" />
        <Skeleton className="h-12 w-full max-w-2xl" />
        <Skeleton className="h-5 w-full max-w-xl" />
      </div>
      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_22rem]">
        <div className="flex flex-col gap-6">
          <Skeleton className="h-52 w-full" />
          <Skeleton className="h-80 w-full" />
        </div>
        <Skeleton className="h-96 w-full" />
      </div>
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

function formatDateRange(startAt: string, endAt: string) {
  return `${formatTournamentDate(startAt)} – ${formatTournamentDate(endAt)}`;
}

function formatMoney(amount: number, currency: string, freeLabel = false) {
  if (freeLabel && amount === 0) return "Free entry";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(amount);
}

function formatPercentage(value: number) {
  return `${new Intl.NumberFormat("en-US", {
    maximumFractionDigits: 1,
  }).format(value)}%`;
}

function initials(name: string) {
  return name
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
}
