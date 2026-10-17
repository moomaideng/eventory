import { AcceptApplicationDialog } from "@/features/organizer/components/accept-application-dialog";
import { RejectApplicationDialog } from "@/features/organizer/components/reject-application-dialog";
import type { DashboardEntry } from "@/features/organizer/utils";

/**
 * The decisions that make sense from each entry status. An application waits
 * for the organizer as LOCKED (a solo player is a one-person team, so the same
 * status applies); forming and payment-pending applicants have nothing to
 * review yet. This only decides which buttons to offer; the server still
 * checks every transition.
 */
export function ReviewActions({
  tournamentId,
  entry,
}: {
  tournamentId: string;
  entry: Pick<DashboardEntry, "id" | "name" | "status">;
}) {
  const awaitingReview = entry.status === "LOCKED";
  const canAccept = awaitingReview || entry.status === "REJECTED";
  const canReject = awaitingReview || entry.status === "ACCEPTED";

  if (!canAccept && !canReject) return null;

  return (
    <div className="flex flex-wrap gap-2">
      {canAccept ? (
        <AcceptApplicationDialog
          tournamentId={tournamentId}
          teamId={entry.id}
          teamName={entry.name}
        />
      ) : null}
      {canReject ? (
        <RejectApplicationDialog
          tournamentId={tournamentId}
          teamId={entry.id}
          teamName={entry.name}
        />
      ) : null}
    </div>
  );
}
