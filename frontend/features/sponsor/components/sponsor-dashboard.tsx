"use client";

import Link from "next/link";
import { useState } from "react";
import { ArrowRight } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
} from "@/components/ui/empty";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

const statusLabels = {
  PENDING: "Pending confirmation",
  CONFIRMED: "Confirmed",
  CANCELLED: "Cancelled",
  EXPIRED: "Expired",
};

export function SponsorDashboard() {
  const { authorizationHeader } = useAuth();
  const [page, setPage] = useState(1);
  const limit = 20;
  const query = $api.useQuery(
    "get",
    "/api/v1/tournaments/sponsorships/mine",
    {
      params: { query: { page, limit } },
      headers: { Authorization: authorizationHeader },
    },
    {
      enabled: Boolean(authorizationHeader),
      refetchInterval: (q) =>
        q.state.data?.items?.some((p) => p.status === "PENDING")
          ? 10000
          : false,
    }
  );

  return (
    <div className="sponsorship-surface mx-auto flex w-full max-w-5xl flex-1 flex-col gap-12 px-4 py-10 sm:px-8 sm:py-14">
      <header className="flex flex-col items-start gap-6 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex max-w-xl flex-col gap-3">
          <h1 className="text-3xl font-bold tracking-tight sm:text-4xl">
            Sponsor dashboard
          </h1>
          <p className="text-muted-foreground text-base leading-relaxed">
            Keep track of your sponsorship packages and the tournaments you
            support.
          </p>
        </div>
        <Button
          render={<Link href="/sponsor/tournaments" />}
          nativeButton={false}
        >
          Find tournaments to sponsor <ArrowRight data-icon="inline-end" />
        </Button>
      </header>

      <section
        aria-labelledby="sponsorships-heading"
        className="flex flex-col gap-5"
      >
        <div className="flex items-center justify-between gap-4">
          <h2 id="sponsorships-heading" className="text-xl font-semibold">
            My sponsorships
          </h2>
          <Button
            variant="outline"
            size="sm"
            onClick={() => void query.refetch()}
            disabled={!authorizationHeader || query.isFetching}
          >
            Refresh
          </Button>
        </div>
        <p className="text-muted-foreground text-sm">
          View your packages and visit their tournament pages. Invoices and
          receipts are kept separately in{" "}
          <Link
            href="/sponsor/payments"
            className="underline underline-offset-4"
          >
            My payments
          </Link>
          .
        </p>
        {query.isPending ? (
          <div
            aria-label="Loading sponsorships"
            className="flex flex-col gap-3"
          >
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-16 w-full" />
          </div>
        ) : query.isError ? (
          <Alert variant="destructive">
            <AlertTitle>Could not load your sponsorships</AlertTitle>
            <AlertDescription>Use Refresh to try again.</AlertDescription>
          </Alert>
        ) : !query.data.items?.length ? (
          <Empty className="border">
            <EmptyHeader>
              <EmptyTitle>
                {page === 1 ? "No sponsorships yet" : "No more sponsorships"}
              </EmptyTitle>
              <EmptyDescription>
                {page === 1
                  ? "Find a tournament and choose a sponsorship package to get started."
                  : "Go back to the previous page to view your packages."}
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Tournament / package</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="hidden text-right sm:table-cell">
                  Tournament
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {query.data.items.map((pledge) => (
                <TableRow key={pledge.id}>
                  <TableCell className="max-w-48 py-5 whitespace-normal sm:max-w-sm">
                    <Link
                      href={`/sponsor/tournaments/${pledge.tournamentId}`}
                      className="font-medium underline-offset-4 hover:underline"
                    >
                      {pledge.tournamentName}
                    </Link>
                    <p className="text-muted-foreground mt-1 text-sm">
                      {pledge.tierName}
                    </p>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={
                        pledge.status === "CONFIRMED"
                          ? "default"
                          : pledge.status === "PENDING"
                            ? "secondary"
                            : "outline"
                      }
                    >
                      {statusLabels[pledge.status]}
                    </Badge>
                  </TableCell>
                  <TableCell className="hidden text-right sm:table-cell">
                    <Link
                      href={`/sponsor/tournaments/${pledge.tournamentId}`}
                      aria-label={`View tournament: ${pledge.tournamentName}`}
                      className="text-sm underline underline-offset-4"
                    >
                      View
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
        {query.data && (page > 1 || query.data.total > limit) && (
          <nav
            aria-label="Sponsorship pages"
            className="flex items-center justify-end gap-3"
          >
            <Button
              variant="outline"
              size="sm"
              disabled={page === 1 || query.isFetching}
              onClick={() => setPage(page - 1)}
            >
              Previous
            </Button>
            <span className="text-muted-foreground text-sm">Page {page}</span>
            <Button
              variant="outline"
              size="sm"
              disabled={page * limit >= query.data.total || query.isFetching}
              onClick={() => setPage(page + 1)}
            >
              Next
            </Button>
          </nav>
        )}
      </section>
    </div>
  );
}
