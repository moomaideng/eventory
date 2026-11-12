"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Mail, Users } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

import { SponsorDirectoryFilters } from "@/features/organizer/components/sponsor-directory-filters";
import { Skeleton } from "@/components/ui/skeleton";

export function SponsorDirectory() {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [page, setPage] = useState(1);
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search.trim()), 250);
    return () => clearTimeout(timer);
  }, [search]);
  function clearFilters() {
    setSearch("");
    setTags([]);
    setPage(1);
  }
  function chooseTag(tag: string) {
    if (!tags.includes(tag) && tags.length < 10) setTags([...tags, tag]);
    setSearch("");
    setPage(1);
  }
  const { data, error, isLoading, isFetching, refetch } = $api.useQuery(
    "get",
    "/api/v1/accounts/sponsors",
    {
      params: {
        query: {
          search: debouncedSearch,
          tag: tags.join(","),
          page,
          pageSize: 12,
        },
      },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
    }
  );

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <header className="space-y-2">
        <p className="text-muted-foreground text-xs font-medium uppercase">
          Organizer workspace
        </p>
        <h1 className="text-2xl font-bold sm:text-3xl">Sponsor directory</h1>
        <p className="text-muted-foreground text-sm">
          Find sponsor profiles and contact them about your tournament.
        </p>
      </header>

      <SponsorDirectoryFilters
        search={search}
        onSearch={(value) => {
          setSearch(value);
          setPage(1);
        }}
        selected={tags}
        onAdd={chooseTag}
        onRemove={(tag) => {
          setTags(tags.filter((item) => item !== tag));
          setPage(1);
        }}
        onClear={clearFilters}
        tags={data?.tags ?? undefined}
        loading={
          authLoading ||
          isLoading ||
          isFetching ||
          search.trim() !== debouncedSearch
        }
        error={Boolean(error)}
      />

      {authLoading || (authorizationHeader && isLoading) ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[0, 1, 2].map((item) => (
            <Skeleton key={item} className="h-40 rounded-xl" />
          ))}
        </div>
      ) : !authorizationHeader ? (
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Users />
            </EmptyMedia>
            <EmptyTitle>Sign in to browse sponsors</EmptyTitle>
            <EmptyDescription>
              Only organizers can view sponsor contact details.
            </EmptyDescription>
          </EmptyHeader>
          <Button
            render={<Link href="/login?redirectTo=%2Forganizer%2Fsponsors" />}
            nativeButton={false}
          >
            Sign in
          </Button>
        </Empty>
      ) : error ? (
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Users />
            </EmptyMedia>
            <EmptyTitle>Could not load sponsors</EmptyTitle>
            <EmptyDescription>
              Check that your organizer profile is ready, then try again.
            </EmptyDescription>
          </EmptyHeader>
          <Button variant="outline" onClick={() => void refetch()}>
            Try again
          </Button>
        </Empty>
      ) : !data?.items?.length ? (
        <Empty className="border">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <Users />
            </EmptyMedia>
            <EmptyTitle>No sponsors found</EmptyTitle>
            <EmptyDescription>
              Sponsors must match every selected tag and your search. Remove a
              tag or try another name.
            </EmptyDescription>
          </EmptyHeader>
          {page > 1 || search || tags.length ? (
            <Button variant="outline" onClick={clearFilters}>
              Clear filters
            </Button>
          ) : null}
        </Empty>
      ) : (
        <>
          <p className="text-muted-foreground text-sm" aria-live="polite">
            {data.total} {data.total === 1 ? "sponsor" : "sponsors"} found
            {isFetching ? " · Updating…" : ""}
          </p>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {(data.items ?? []).map((sponsor) => (
              <Card key={sponsor.accountId} id={`sponsor-${sponsor.accountId}`}>
                <CardHeader>
                  <CardTitle className="text-lg">
                    {sponsor.sponsorName}
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex flex-wrap gap-2">
                    {(sponsor.tags ?? []).map((tag) => (
                      <Button
                        key={tag}
                        variant="secondary"
                        size="sm"
                        disabled={tags.includes(tag) || tags.length >= 10}
                        onClick={() => chooseTag(tag)}
                        aria-label={`Filter by ${tag}`}
                      >
                        #{tag}
                      </Button>
                    ))}
                  </div>
                  <a
                    className="text-primary inline-flex items-center gap-2 text-sm font-medium hover:underline"
                    href={`mailto:${sponsor.contactEmail}`}
                  >
                    <Mail className="size-4" aria-hidden="true" />
                    {sponsor.contactEmail}
                  </a>
                </CardContent>
              </Card>
            ))}
          </div>
          <div className="flex items-center justify-end gap-3">
            <Button
              variant="outline"
              disabled={page === 1 || isFetching}
              onClick={() => setPage(page - 1)}
            >
              Previous
            </Button>
            <span className="text-muted-foreground text-sm">
              Page {page} of{" "}
              {Math.max(1, Math.ceil(data.total / data.pageSize))}
            </span>
            <Button
              variant="outline"
              disabled={page * data.pageSize >= data.total || isFetching}
              onClick={() => setPage(page + 1)}
            >
              Next
            </Button>
          </div>
        </>
      )}
    </div>
  );
}
