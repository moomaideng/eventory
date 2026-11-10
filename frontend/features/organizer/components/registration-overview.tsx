"use client";

import React from "react";
import {
  ReceiptText,
  ShieldCheck,
  Timer,
  Users,
} from "lucide-react";
import { Progress } from "@/components/ui/progress";
import type { OrganizerSummary } from "@/features/organizer/utils";
import {
  formatMoney,
  formatPercentage,
  formatTournamentDate,
} from "@/features/tournaments/utils";

const numberFormatter = new Intl.NumberFormat("en-US");

function getTimeRemaining(deadlineStr: string | undefined | null, status: string, now: number) {
  switch (status) {
    case "COMPLETED":
      return { value: "Completed", detail: "Tournament ended" };
    case "ONGOING":
      return { value: "In progress", detail: "Tournament underway" };
  }

  if (status === "DRAFT") {
    return {
      value: "Registration not open",
      detail: deadlineStr
        ? `Opens when you start registration • deadline ${formatTournamentDate(deadlineStr)}`
        : "Opens when you start registration",
    };
  }
  if (!deadlineStr) {
    return { value: "Not set", detail: "No deadline configured" };
  }

  const deadline = new Date(deadlineStr).getTime();
  const diffMs = deadline - now;
  if (diffMs <= 0) {
    return status === "CROWDFUNDING"
      ? {
          value: "Funding deadline passed",
          detail: "Registration is still closed",
        }
      : { value: "Registration closed", detail: "Passed deadline" };
  }
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffHours / 24);
  if (diffDays > 1) {
    return {
      value: `${diffDays} days left`,
      detail: `Closes ${formatTournamentDate(deadlineStr)}`,
    };
  }
  if (diffDays === 1) {
    return {
      value: "1 day left",
      detail: `Closes ${formatTournamentDate(deadlineStr)}`,
    };
  }
  if (diffHours > 0) {
    return {
      value: `${diffHours} hours left`,
      detail: `Closes ${formatTournamentDate(deadlineStr)}`,
    };
  }
  const diffMins = Math.max(1, Math.floor(diffMs / (1000 * 60)));
  return {
    value: `${diffMins} mins left`,
    detail: "Closing soon",
  };
}

export function RegistrationOverview({ summary }: { summary: OrganizerSummary }) {
  const [now] = React.useState(() => Date.now());
  const { tournament, metrics, funding } = summary;
  const entryUnit =
    tournament.registrationMode === "TEAM"
      ? "team entries"
      : tournament.registrationMode === "SOLO"
        ? "solo entries"
        : "entries";

  const deadline =
    tournament.status === "CROWDFUNDING" && funding.deadline
      ? funding.deadline
      : tournament.registrationDeadline;

  const timeRemaining = getTimeRemaining(deadline, tournament.status, now);

  const pipelineDetails = [
    `${metrics.formingEntries} forming`,
    metrics.lockedEntries > 0
      ? `${metrics.lockedEntries} locked awaiting review`
      : "0 locked awaiting review",
    `${metrics.acceptedEntries} accepted`,
  ].join(" • ");

  const organizerMetrics = [
    {
      label: "Tournament capacity",
      value:
        tournament.capacity > 0
          ? `${metrics.acceptedEntries} / ${tournament.capacity}`
          : `${metrics.acceptedEntries} accepted`,
      detail: `${metrics.confirmedParticipants} confirmed players • ${tournament.capacity > 0 ? `${metrics.availableSpots} ${entryUnit} left` : "Capacity open"}`,
      icon: Users,
    },
    {
      label: "Registration pipeline",
      value: `${metrics.totalEntries} ${metrics.totalEntries === 1 ? "entry" : "entries"}`,
      detail: pipelineDetails,
      icon: ShieldCheck,
    },
    {
      label:
        tournament.status === "CROWDFUNDING"
          ? "Funding time left"
          : "Time remaining",
      value: timeRemaining.value,
      detail: timeRemaining.detail,
      icon: Timer,
    },
  ];

  const progressValue = Math.min(Math.max(funding.percentage, 0), 100);
  const hasFunding = funding.goalAmount > 0;

  return (
    <div className="flex flex-col gap-6">
      {/* Top KPI Metrics */}
      <dl className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {organizerMetrics.map(({ label, value, detail, icon: Icon }) => (
          <div
            key={label}
            className="border-border bg-card/50 flex flex-col justify-between gap-3 rounded-xl border p-5 shadow-xs"
          >
            <div className="flex items-center justify-between">
              <dt className="text-muted-foreground text-xs font-medium uppercase tracking-wider">
                {label}
              </dt>
              <div className="bg-primary/10 text-primary flex size-8 items-center justify-center rounded-md">
                <Icon className="size-4" aria-hidden="true" />
              </div>
            </div>
            <div>
              <dd className="text-2xl font-bold tracking-tight">{value}</dd>
              <p className="text-muted-foreground mt-1 text-xs">{detail}</p>
            </div>
          </div>
        ))}
      </dl>

      {/* Funding Progress (if active/configured) */}
      {hasFunding ? (
        <div className="border-border bg-card/50 rounded-xl border p-6 shadow-xs sm:p-7">
          <div className="flex items-center justify-between gap-4 border-b pb-4">
            <div className="flex items-center gap-2.5">
              <ReceiptText className="text-primary size-5" aria-hidden="true" />
              <h3 className="text-base font-semibold">Crowdfunding Campaign</h3>
            </div>
            <span className="text-muted-foreground text-xs font-medium">
              {funding.deadline
                ? `Deadline: ${formatTournamentDate(funding.deadline)}`
                : "No deadline"}
            </span>
          </div>

          <div className="mt-5 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
            <div className="flex flex-col gap-1">
              <span className="text-muted-foreground text-xs font-medium uppercase">
                Raised Amount
              </span>
              <span className="text-2xl font-bold tracking-tight">
                {formatMoney(funding.raisedAmount, funding.currency)}
              </span>
              <span className="text-muted-foreground text-xs">
                Goal: {formatMoney(funding.goalAmount, funding.currency)}
              </span>
            </div>

            <div className="flex flex-col gap-1.5 sm:col-span-2">
              <div className="text-muted-foreground flex items-center justify-between text-xs">
                <span className="font-medium uppercase">Campaign Progress</span>
                <span className="font-semibold">{formatPercentage(funding.percentage)}</span>
              </div>
              <Progress value={progressValue} className="h-2" />
              <span className="text-muted-foreground text-xs">
                {formatMoney(funding.remainingAmount, funding.currency)} remaining to reach goal
              </span>
            </div>

            <div className="flex flex-col gap-1">
              <span className="text-muted-foreground text-xs font-medium uppercase">
                Supporters
              </span>
              <span className="text-2xl font-bold tracking-tight">
                {numberFormatter.format(funding.supporterCount)}
              </span>
              <span className="text-muted-foreground text-xs">Direct pledges</span>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
