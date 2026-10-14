import { AcceptApplicationDialog } from "@/features/organizer/components/accept-application-dialog";
import { RejectApplicationDialog } from "@/features/organizer/components/reject-application-dialog";
import type { DashboardEntry } from "@/features/organizer/utils";

/**
 * The decisions that make sense from each entry status. Forming and
 * payment-pending teams have nothing to review yet. This only decides which
 * buttons to offer; the server still checks every transition.
 */
export function ReviewActions({
  tournamentId,
  entry,
}: {
  tournamentId: string;
  entry: Pick<DashboardEntry, "id" | "name" | "status">;
}) {
  const canAccept = entry.status === "LOCKED" || entry.status === "REJECTED";
  const canReject = entry.status === "LOCKED" || entry.status === "ACCEPTED";

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
