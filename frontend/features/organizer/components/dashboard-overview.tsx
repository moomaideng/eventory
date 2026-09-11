import {
  CalendarDays,
  Clock3,
  MapPin,
  ReceiptText,
  ShieldCheck,
  Timer,
  Users,
  WalletCards,
} from "lucide-react";
import { Progress } from "@/components/ui/progress";
import type { OrganizerSummary } from "@/features/organizer/utils";
import {
  formatDateRange,
  formatMoney,
  formatPercentage,
  formatRegistrationType,
  formatTournamentDate,
} from "@/features/tournaments/utils";

const numberFormatter = new Intl.NumberFormat("en-US");

function getTimeRemaining(deadlineStr: string, status: string) {
  if (status === "COMPLETED") {
    return { value: "Completed", detail: "Tournament ended" };
  }
  if (status === "IN_PROGRESS") {
    return { value: "In progress", detail: "Tournament underway" };
  }
  const now = Date.now();
  const deadline = new Date(deadlineStr).getTime();
  const diffMs = deadline - now;
  if (diffMs <= 0) {
    return { value: "Registration closed", detail: "Passed deadline" };
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

export function DashboardOverview({ summary }: { summary: OrganizerSummary }) {
  const { tournament, metrics, funding } = summary;
  const entryUnit =
    tournament.registrationMode === "TEAM"
      ? "team entries"
      : tournament.registrationMode === "SOLO"
        ? "solo entries"
        : "entries";

  const timeRemaining = getTimeRemaining(
    tournament.registrationDeadline,
    tournament.status
  );

  const readinessDetail =
    metrics.formingEntries > 0
      ? `${metrics.formingEntries} ${metrics.formingEntries === 1 ? "team" : "teams"} still forming`
      : metrics.acceptedEntries > 0
        ? "All rosters locked"
        : "No accepted teams yet";

  const organizerMetrics = [
    {
      label: "Tournament capacity",
      value: `${metrics.acceptedEntries} / ${tournament.capacity}`,
      detail: `${metrics.confirmedParticipants} confirmed players • ${metrics.availableSpots} ${entryUnit} left`,
      icon: Users,
    },
    {
      label: "Roster readiness",
      value: `${metrics.lockedEntries} / ${metrics.acceptedEntries} Locked`,
      detail: readinessDetail,
      icon: ShieldCheck,
    },
    {
      label: "Time remaining",
      value: timeRemaining.value,
      detail: timeRemaining.detail,
      icon: Timer,
    },
  ];

  const overviewItems = [
    {
      icon: CalendarDays,
      label: "Tournament schedule",
      value: formatDateRange(tournament.startAt, tournament.endAt),
    },
    {
      icon: Clock3,
      label: "Registration deadline",
      value: formatTournamentDate(tournament.registrationDeadline),
    },
    {
      icon: MapPin,
      label: "Location",
      value: tournament.location,
    },
    {
      icon: WalletCards,
      label: "Entry fee",
      value: formatMoney(tournament.entryFee, tournament.currency, true),
    },
    {
      icon: Users,
      label: "Availability",
      value: `${metrics.availableSpots} of ${tournament.capacity} spots available`,
    },
    {
      icon: Users,
      label: "Registration type",
      value: formatRegistrationType(tournament.registrationMode),
    },
  ];

  const progressValue = Math.min(Math.max(funding.percentage, 0), 100);

  return (
    <div className="flex flex-col gap-10">
      {/* 3 Top KPI Cards */}
      <dl className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {organizerMetrics.map(({ label, value, detail, icon: Icon }) => (
          <div
            key={label}
            className="bg-card flex min-w-0 flex-col gap-3 rounded-xl border border-border/50 p-4 shadow-xs sm:p-5"
          >
            <div className="bg-primary/10 text-primary flex size-8 items-center justify-center rounded-lg">
              <Icon className="size-4" aria-hidden="true" />
            </div>
            <dt className="text-muted-foreground text-xs sm:text-sm">{label}</dt>
            <dd className="flex flex-col gap-0.5">
              <span className="text-2xl font-bold tracking-tight tabular-nums sm:text-3xl">
                {value}
              </span>
              <span className="text-muted-foreground text-xs">{detail}</span>
            </dd>
          </div>
        ))}
      </dl>

      {/* Middle Zone: Event overview & Funding in a distinct surface container */}
      <div className="bg-card/30 rounded-2xl border border-border/50 p-6 shadow-xs sm:p-8">
        <div className="grid items-start gap-8 lg:grid-cols-[minmax(0,1fr)_22rem] lg:gap-10">
          {/* Left Column: Event overview */}
          <div className="flex flex-col gap-5">
            <div className="flex items-center gap-2 text-base font-semibold">
              <CalendarDays className="size-4 text-muted-foreground" aria-hidden="true" />
              <h2>Event overview</h2>
            </div>

            {tournament.description ? (
              <p className="text-muted-foreground text-sm leading-relaxed">
                {tournament.description}
              </p>
            ) : null}

            <div className="grid gap-5 sm:grid-cols-2">
              {overviewItems.map(({ icon: Icon, label, value }) => (
                <div key={label} className="flex items-start gap-3">
                  <Icon className="text-muted-foreground mt-0.5 size-5 shrink-0" />
                  <div className="flex flex-col gap-1">
                    <p className="text-sm font-medium">{label}</p>
                    <p className="text-muted-foreground text-sm wrap-anywhere">
                      {value}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Right Column: Funding (with subtle vertical divider on desktop) */}
          <div className="flex flex-col gap-5 lg:border-l lg:border-border/40 lg:pl-10">
            <div className="flex items-center gap-2 text-base font-semibold">
              <ReceiptText className="size-4 text-muted-foreground" aria-hidden="true" />
              <h2>Funding</h2>
            </div>

            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-0.5">
                <span className="text-3xl font-bold tracking-tight">
                  {formatMoney(funding.raisedAmount, funding.currency)}
                </span>
                <span className="text-muted-foreground text-xs">
                  raised of {formatMoney(funding.goalAmount, funding.currency)}
                </span>
              </div>

              <div className="flex flex-col gap-1.5">
                <div className="flex items-center justify-between text-xs text-muted-foreground">
                  <span>Funding goal</span>
                  <span>{formatPercentage(funding.percentage)}</span>
                </div>
                <Progress value={progressValue} className="h-1.5" />
              </div>

              <div className="grid grid-cols-2 gap-4 border-t border-border/40 pt-3">
                <div className="flex flex-col gap-1">
                  <span className="text-muted-foreground text-xs">Supporters</span>
                  <span className="text-sm font-semibold">
                    {numberFormatter.format(funding.supporterCount)}
                  </span>
                </div>
                <div className="flex flex-col gap-1">
                  <span className="text-muted-foreground text-xs">Still needed</span>
                  <span className="text-sm font-semibold">
                    {formatMoney(funding.remainingAmount, funding.currency)}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
