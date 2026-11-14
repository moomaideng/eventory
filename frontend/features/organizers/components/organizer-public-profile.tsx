"use client";

import React from "react";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import {
  ArrowLeft,
  ArrowRight,
  CalendarDays,
  Mail,
  Phone,
  Trophy,
} from "lucide-react";
import { $api } from "@/lib/api/client";
import { PageHeader } from "@/components/page-header";
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
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { TournamentGridSkeleton } from "@/features/tournaments/components/tournament-grid-skeleton";
import { PAGE_SIZE, formatDate, initials } from "@/features/tournaments/utils";
import {
  organizerTournamentsQuerySchema,
  type OrganizerTournamentStatusFilter,
} from "../schemas";
import { OrganizerProfileSkeleton } from "./organizer-profile-skeleton";

const STATUS_TABS: ReadonlyArray<{
  value: OrganizerTournamentStatusFilter;
  label: string;
}> = [
  { value: "ALL", label: "All" },
  { value: "CROWDFUNDING", label: "Crowdfunding" },
  { value: "REGISTRATION", label: "Registration" },
  { value: "ONGOING", label: "Ongoing" },
  { value: "COMPLETED", label: "Completed" },
];

export function OrganizerPublicProfile({
  organizerId,
}: {
  organizerId: string;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const parsedQuery = organizerTournamentsQuerySchema.safeParse({
    status: searchParams.get("status") ?? undefined,
    page: searchParams.get("page") ?? undefined,
  });

  const { status, page } = parsedQuery.success
    ? parsedQuery.data
    : { status: "ALL" as const, page: 1 };

  const profileQuery = $api.useQuery(
    "get",
    "/api/v1/accounts/organizers/{id}",
    { params: { path: { id: organizerId } } },
    { staleTime: 30_000, retry: false }
  );

  const tournamentsQuery = $api.useQuery(
    "get",
    "/api/v1/tournaments",
    {
      params: {
        query: {
          organizerId,
          status: status === "ALL" ? undefined : status,
          sort: "start_desc",
          page,
          pageSize: PAGE_SIZE,
        },
      },
    },
    { staleTime: 30_000 }
  );

  function updateQuery(nextStatus: OrganizerTournamentStatusFilter, nextPage: number) {
    const next = new URLSearchParams();
    if (nextStatus !== "ALL") {
      next.set("status", nextStatus);
    }
    if (nextPage > 1) {
      next.set("page", String(nextPage));
    }
    router.push(next.size ? `${pathname}?${next.toString()}` : pathname);
  }

  function handleStatusChange(value: unknown) {
    const parsed = organizerTournamentsQuerySchema.shape.status.safeParse(value);
    if (!parsed.success || parsed.data === status) {
      return;
    }
    updateQuery(parsed.data, 1);
  }

  if (profileQuery.isLoading) {
    return <OrganizerProfileSkeleton />;
  }

  if (profileQuery.error || !profileQuery.data) {
    return (
      <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-6 px-4 py-10 sm:px-8">
        <PageHeader
          breadcrumbs={[
            { label: "Tournaments", href: "/tournaments" },
            { label: "Organizer unavailable" },
          ]}
          title="Organizer unavailable"
        />
        <Alert variant="destructive">
          <AlertTitle>Organizer profile unavailable</AlertTitle>
          <AlertDescription>
            This organizer profile may not exist or the API may be temporarily
            unavailable.
          </AlertDescription>
        </Alert>
        <div>
          <Button
            render={<Link href="/tournaments" />}
            nativeButton={false}
            variant="outline"
          >
            <ArrowLeft data-icon="inline-start" />
            Back to tournaments
          </Button>
        </div>
      </div>
    );
  }

  const profile = profileQuery.data;
  const tournamentsData = tournamentsQuery.data;
  const items = tournamentsData?.items ?? [];
  const pageCount = tournamentsData
    ? Math.max(1, Math.ceil(tournamentsData.total / tournamentsData.pageSize))
    : 1;

  return (
    <div className="container mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        breadcrumbs={[
          { label: "Tournaments", href: "/tournaments" },
          { label: profile.organizerName },
        ]}
        title={profile.organizerName}
        description="Organizer profile and hosted tournaments."
        actions={
          <Button
            render={<Link href="/tournaments" />}
            nativeButton={false}
            variant="outline"
          >
            Browse all tournaments
            <ArrowRight data-icon="inline-end" />
          </Button>
        }
      />

      <Card>
        <CardHeader>
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex min-w-0 items-center gap-4">
              <Avatar size="lg">
                <AvatarFallback>
                  {initials(profile.organizerName) || "OR"}
                </AvatarFallback>
              </Avatar>
              <div className="flex min-w-0 flex-col gap-1">
                <CardTitle className="truncate">{profile.organizerName}</CardTitle>
                <CardDescription>Tournament Organizer</CardDescription>
              </div>
            </div>
            <Badge variant="secondary">Public organizer profile</Badge>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <Separator />
          <dl className="grid gap-4 sm:grid-cols-3">
            <div className="flex items-start gap-3">
              <Mail
                aria-hidden="true"
                className="text-muted-foreground mt-0.5 size-4 shrink-0"
              />
              <div className="flex min-w-0 flex-col gap-1">
                <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wide">
                  Contact email
                </dt>
                <dd className="text-sm font-medium wrap-anywhere">
                  {profile.organizerEmail ? (
                    <a
                      href={`mailto:${profile.organizerEmail}`}
                      className="hover:underline underline-offset-4"
                    >
                      {profile.organizerEmail}
                    </a>
                  ) : (
                    <span className="text-muted-foreground font-normal">
                      Not provided
                    </span>
                  )}
                </dd>
              </div>
            </div>

            {profile.phone ? (
              <div className="flex items-start gap-3">
                <Phone
                  aria-hidden="true"
                  className="text-muted-foreground mt-0.5 size-4 shrink-0"
                />
                <div className="flex min-w-0 flex-col gap-1">
                  <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wide">
                    Phone
                  </dt>
                  <dd className="text-sm font-medium wrap-anywhere">
                    <a
                      href={`tel:${profile.phone}`}
                      className="hover:underline underline-offset-4"
                    >
                      {profile.phone}
                    </a>
                  </dd>
                </div>
              </div>
            ) : null}

            <div className="flex items-start gap-3">
              <CalendarDays
                aria-hidden="true"
                className="text-muted-foreground mt-0.5 size-4 shrink-0"
              />
              <div className="flex min-w-0 flex-col gap-1">
                <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wide">
                  Member since
                </dt>
                <dd className="text-sm font-medium">
                  {formatDate(profile.createdAt)}
                </dd>
              </div>
            </div>
          </dl>
        </CardContent>
      </Card>

      <section
        aria-labelledby="hosted-tournaments-heading"
        className="flex flex-col gap-5"
      >
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-baseline gap-3">
            <h2
              id="hosted-tournaments-heading"
              className="font-heading text-xl font-semibold tracking-tight"
            >
              Hosted tournaments
            </h2>
            {tournamentsData ? (
              <span className="text-muted-foreground text-sm">
                {tournamentsData.total} tournament
                {tournamentsData.total === 1 ? "" : "s"}
              </span>
            ) : null}
          </div>
          {tournamentsQuery.isFetching && !tournamentsQuery.isLoading ? (
            <span className="text-muted-foreground text-sm">Updating…</span>
          ) : null}
        </div>

        <Tabs value={status} onValueChange={handleStatusChange}>
          <TabsList className="max-w-full justify-start overflow-x-auto">
            {STATUS_TABS.map((tab) => (
              <TabsTrigger key={tab.value} value={tab.value}>
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>

          <TabsContent value={status}>
            {tournamentsQuery.error ? (
              <Alert variant="destructive">
                <AlertTitle>Could not load hosted tournaments</AlertTitle>
                <AlertDescription>
                  Please try again after the API is available.
                </AlertDescription>
              </Alert>
            ) : tournamentsQuery.isLoading ? (
              <TournamentGridSkeleton />
            ) : items.length === 0 ? (
              <Empty className="border">
                <EmptyHeader>
                  <EmptyMedia variant="icon">
                    <Trophy />
                  </EmptyMedia>
                  <EmptyTitle>
                    {status === "ALL"
                      ? "No hosted tournaments yet"
                      : "No tournaments in this stage"}
                  </EmptyTitle>
                  <EmptyDescription>
                    {status === "ALL"
                      ? `${profile.organizerName} has not published any tournaments yet.`
                      : `${profile.organizerName} does not have any published tournaments with this status.`}
                  </EmptyDescription>
                </EmptyHeader>
                <EmptyContent>
                  {status !== "ALL" ? (
                    <Button
                      variant="outline"
                      onClick={() => updateQuery("ALL", 1)}
                    >
                      Show all tournaments
                    </Button>
                  ) : (
                    <Button
                      render={<Link href="/tournaments" />}
                      nativeButton={false}
                      variant="outline"
                    >
                      Browse all tournaments
                    </Button>
                  )}
                </EmptyContent>
              </Empty>
            ) : (
              <div className="flex flex-col gap-6">
                <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
                  {items.map((tournament) => (
                    <TournamentCard key={tournament.id} tournament={tournament} />
                  ))}
                </div>

                {tournamentsData && tournamentsData.total > tournamentsData.pageSize ? (
                  <div className="flex items-center justify-center gap-3">
                    <Button
                      variant="outline"
                      disabled={page <= 1 || tournamentsQuery.isFetching}
                      onClick={() => updateQuery(status, page - 1)}
                    >
                      <ArrowLeft data-icon="inline-start" />
                      Previous
                    </Button>
                    <span className="text-muted-foreground text-sm">
                      Page {page} of {pageCount}
                    </span>
                    <Button
                      variant="outline"
                      disabled={page >= pageCount || tournamentsQuery.isFetching}
                      onClick={() => updateQuery(status, page + 1)}
                    >
                      Next
                      <ArrowRight data-icon="inline-end" />
                    </Button>
                  </div>
                ) : null}
              </div>
            )}
          </TabsContent>
        </Tabs>
      </section>
    </div>
  );
}
