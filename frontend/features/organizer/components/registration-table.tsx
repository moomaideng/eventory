"use client";

import { useState } from "react";
import { ClipboardList } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
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
  TableCaption,
} from "@/components/ui/table";
import {
  DashboardPagination,
  DashboardStatus,
} from "@/features/organizer/components/dashboard-shared";
import type {
  DashboardEntry,
  OrganizerSummary,
} from "@/features/organizer/utils";
import { formatDate, initials } from "@/features/tournaments/utils";

export function RegistrationTable({
  entries,
  metrics,
}: {
  entries: DashboardEntry[];
  metrics: OrganizerSummary["metrics"];
}) {
  const [requestedPage, setPage] = useState(1);
  const pages = Math.max(1, Math.ceil(entries.length / 10));
  const page = Math.min(requestedPage, pages);
  const visible = entries.slice((page - 1) * 10, page * 10);
  const counts = [
    ["Total entries", metrics.totalEntries],
    ["Forming", metrics.formingEntries],
    ["Locked", metrics.lockedEntries],
    ["Accepted", metrics.acceptedEntries],
    ["Rejected", metrics.rejectedEntries],
  ] as const;
  return (
    <section
      className="flex min-w-0 flex-col gap-5"
      aria-labelledby="registrations"
    >
      <div className="flex items-center justify-between gap-3">
        <h2
          id="registrations"
          className="flex items-center gap-2 text-lg font-semibold"
        >
          <ClipboardList className="size-5" />
          Registrations
        </h2>
        <Badge variant="secondary">
          {entries.length} {entries.length === 1 ? "entry" : "entries"}
        </Badge>
      </div>
      <dl className="grid grid-cols-2 gap-4 border-y py-4 sm:grid-cols-5">
        {counts.map(([label, count]) => (
          <div key={label}>
            <dt className="text-muted-foreground text-xs">{label}</dt>
            <dd className="mt-1 font-semibold tabular-nums">{count}</dd>
          </div>
        ))}
      </dl>
      {!entries.length ? (
        <Empty className="border">
          <EmptyHeader>
            <EmptyTitle>No registrations yet</EmptyTitle>
            <EmptyDescription>
              Entries will appear when participants form a team or register.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="overflow-hidden rounded-lg border">
          <Table className="table-fixed">
            <TableCaption className="sr-only">
              Tournament entries, status, and participant rosters
            </TableCaption>
            <TableHeader>
              <TableRow>
                <TableHead className="w-1/3 sm:w-1/4">Entry</TableHead>
                <TableHead className="hidden w-32 sm:table-cell">
                  Status
                </TableHead>
                <TableHead>Participants</TableHead>
                <TableHead className="hidden w-32 lg:table-cell">
                  Created
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {visible.map((entry) => (
                <TableRow key={entry.id}>
                  <TableCell className="align-top whitespace-normal">
                    <div className="flex items-start gap-3 py-1">
                      <Avatar className="hidden shrink-0 sm:flex">
                        <AvatarFallback>{initials(entry.name)}</AvatarFallback>
                      </Avatar>
                      <div className="flex min-w-0 flex-col gap-2">
                        <span className="font-medium [overflow-wrap:anywhere]">
                          {entry.name}
                        </span>
                        <span className="sm:hidden">
                          <DashboardStatus status={entry.status} />
                        </span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell className="hidden align-top sm:table-cell">
                    <DashboardStatus status={entry.status} />
                  </TableCell>
                  <TableCell className="align-top whitespace-normal">
                    <ul className="flex flex-col gap-2 py-1">
                      {entry.members?.length ? (
                        entry.members.map((member) => (
                          <li key={member.id} className="min-w-0">
                            <p className="font-medium [overflow-wrap:anywhere]">
                              {member.displayName || member.handle}
                              {member.role === "CAPTAIN" ? (
                                <span className="text-muted-foreground ml-2 text-xs font-normal">
                                  Captain
                                </span>
                              ) : null}
                            </p>
                            <p className="text-muted-foreground text-xs [overflow-wrap:anywhere]">
                              @{member.handle}
                            </p>
                          </li>
                        ))
                      ) : (
                        <li className="text-muted-foreground">No members</li>
                      )}
                    </ul>
                  </TableCell>
                  <TableCell className="text-muted-foreground hidden align-top lg:table-cell">
                    {formatDate(entry.createdAt)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
      <DashboardPagination page={page} pages={pages} onPage={setPage} />
    </section>
  );
}
