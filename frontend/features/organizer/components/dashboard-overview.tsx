import {
  CalendarDays,
  CheckCheck,
  LockKeyhole,
  UserRoundPlus,
  Users,
  Wallet,
} from "lucide-react";
import {
  Progress,
  ProgressLabel,
  ProgressValue,
} from "@/components/ui/progress";
import type { OrganizerSummary } from "@/features/organizer/utils";
import {
  formatMoney,
  formatTournamentDate,
} from "@/features/tournaments/utils";

export function DashboardOverview({ summary }: { summary: OrganizerSummary }) {
  const { tournament: t, metrics: m, funding: f } = summary;
  const entryUnit =
    t.registrationMode === "TEAM"
      ? "team entries"
      : t.registrationMode === "SOLO"
        ? "solo entries"
        : "entries";
  const metrics = [
    {
      label: "Accepted entries",
      value: `${m.acceptedEntries} / ${t.capacity}`,
      detail: entryUnit,
      icon: CheckCheck,
    },
    {
      label: "Confirmed participants",
      value: m.confirmedParticipants,
      detail: "On accepted entries",
      icon: Users,
    },
    {
      label: "Locked entries",
      value: m.lockedEntries,
      detail: "Roster locked",
      icon: LockKeyhole,
    },
    {
      label: "Available spots",
      value: m.availableSpots,
      detail: entryUnit,
      icon: UserRoundPlus,
    },
  ];
  const facts = [
    ["Starts (Bangkok)", formatTournamentDate(t.startAt)],
    ["Ends (Bangkok)", formatTournamentDate(t.endAt)],
    ["Registration closes", formatTournamentDate(t.registrationDeadline)],
    ["Location", t.location],
    ["Entry fee", formatMoney(t.entryFee, t.currency, true)],
    [
      "Registration",
      `${t.registrationMode === "TEAM" ? "Team" : t.registrationMode === "SOLO" ? "Solo" : "Solo or team"} / ${t.minTeamSize}-${t.maxTeamSize} players`,
    ],
  ];
  return (
    <>
      <dl className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        {metrics.map(({ label, value, detail, icon: Icon }) => (
          <div
            key={label}
            className="bg-card flex min-w-0 flex-col gap-3 rounded-lg border p-4 shadow-xs sm:p-5"
          >
            <div className="bg-muted text-primary flex size-9 items-center justify-center rounded-md">
              <Icon className="size-4" aria-hidden="true" />
            </div>
            <dt className="text-muted-foreground min-h-10 text-sm">{label}</dt>
            <dd className="flex flex-col gap-1">
              <span className="text-2xl font-semibold tabular-nums">
                {value}
              </span>
              <span className="text-muted-foreground text-xs">{detail}</span>
            </dd>
          </div>
        ))}
      </dl>
      <div className="bg-muted/20 grid gap-8 border-y px-4 py-6 sm:px-6 lg:grid-cols-[1.5fr_1fr]">
        <section className="min-w-0" aria-labelledby="event-overview">
          <h2
            id="event-overview"
            className="mb-5 flex items-center gap-2 text-lg font-semibold"
          >
            <CalendarDays className="size-5" />
            Event overview
          </h2>
          <dl className="grid gap-x-6 gap-y-5 sm:grid-cols-2">
            {facts.map(([label, value]) => (
              <div key={label} className="min-w-0">
                <dt className="text-muted-foreground text-sm">{label}</dt>
                <dd className="mt-1 text-sm font-medium wrap-anywhere">
                  {value}
                </dd>
              </div>
            ))}
          </dl>
        </section>
        <section
          className="flex min-w-0 flex-col gap-5"
          aria-labelledby="dashboard-funding"
        >
          <h2
            id="dashboard-funding"
            className="flex items-center gap-2 text-lg font-semibold"
          >
            <Wallet className="size-5" />
            Funding
          </h2>
          <p className="text-2xl font-semibold wrap-anywhere">
            {formatMoney(f.raisedAmount, f.currency)}
          </p>
          {f.goalAmount > 0 ? (
            <Progress value={Math.min(100, Math.max(0, f.percentage))}>
              <ProgressLabel>
                {formatMoney(f.goalAmount, f.currency)} goal
              </ProgressLabel>
              <ProgressValue>
                {() =>
                  `${f.percentage.toLocaleString("en-US", { maximumFractionDigits: 1 })}%`
                }
              </ProgressValue>
            </Progress>
          ) : (
            <p className="text-muted-foreground text-sm">No funding goal set</p>
          )}
          <dl className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <dt className="text-muted-foreground">Still needed</dt>
              <dd className="mt-1 font-medium wrap-anywhere">
                {formatMoney(f.remainingAmount, f.currency)}
              </dd>
            </div>
            <div>
              <dt className="text-muted-foreground">Supporters</dt>
              <dd className="mt-1 font-medium tabular-nums">
                {f.supporterCount}
              </dd>
            </div>
          </dl>
        </section>
      </div>
    </>
  );
}
