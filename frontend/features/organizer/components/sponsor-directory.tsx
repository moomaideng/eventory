"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { Mail, Search, Users } from "lucide-react";
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
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";

export function SponsorDirectory() {
  const { authorizationHeader, isLoading: authLoading } = useAuth();
  const [searchDraft, setSearchDraft] = useState("");
  const [categoryDraft, setCategoryDraft] = useState("__all__");
  const [search, setSearch] = useState("");
  const [category, setCategory] = useState("");
  const [page, setPage] = useState(1);
  const { data, error, isLoading, isFetching, refetch } = $api.useQuery(
    "get",
    "/api/v1/accounts/sponsors",
    {
      params: { query: { search, category, page, pageSize: 12 } },
      headers: { Authorization: authorizationHeader ?? "" },
    },
    { enabled: Boolean(authorizationHeader), retry: false }
  );

  function applyFilters(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPage(1);
    setSearch(searchDraft.trim());
    setCategory(categoryDraft === "__all__" ? "" : categoryDraft);
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <header className="space-y-2">
        <p className="text-muted-foreground text-xs font-medium uppercase">
          Organizer workspace
        </p>
        <h1 className="text-2xl font-bold sm:text-3xl">Sponsor directory</h1>
        <p className="text-muted-foreground text-sm">
          Find sponsors who chose to share their contact details with
          organizers.
        </p>
      </header>

      <form
        className="grid gap-3 sm:grid-cols-[1fr_1fr_auto]"
        onSubmit={applyFilters}
      >
        <div>
          <label
            htmlFor="sponsor-search"
            className="mb-1 block text-sm font-medium"
          >
            Sponsor name
          </label>
          <Input
            id="sponsor-search"
            value={searchDraft}
            onChange={(event) => setSearchDraft(event.target.value)}
            placeholder="Search sponsors"
            maxLength={100}
          />
        </div>
        <div>
          <label
            htmlFor="sponsor-category-filter"
            className="mb-1 block text-sm font-medium"
          >
            Category
          </label>
          <Select
            items={[
              { label: "All categories", value: "__all__" },
              ...(data?.categories ?? []).map((value) => ({
                label: value,
                value,
              })),
            ]}
            value={categoryDraft}
            onValueChange={(value) => setCategoryDraft(String(value))}
          >
            <SelectTrigger id="sponsor-category-filter" className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="__all__">All categories</SelectItem>
              {(data?.categories ?? []).map((value) => (
                <SelectItem key={value} value={value}>
                  {value}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Button className="self-end" type="submit">
          <Search data-icon="inline-start" />
          Search
        </Button>
      </form>

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
              Try a different name or category. Sponsors appear here only after
              they opt in.
            </EmptyDescription>
          </EmptyHeader>
          {page > 1 || search || category ? (
            <Button
              variant="outline"
              onClick={() => {
                setPage(1);
                setSearch("");
                setCategory("");
                setSearchDraft("");
                setCategoryDraft("__all__");
              }}
            >
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
              <Card key={sponsor.accountId}>
                <CardHeader>
                  <CardTitle className="text-lg">
                    {sponsor.sponsorName}
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-muted-foreground text-sm">
                    {sponsor.category}
                  </p>
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
