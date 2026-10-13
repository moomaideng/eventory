"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { teamPaymentReferenceSchema } from "@/features/payments/schemas";
import type { Payment } from "@/features/payments/utils";

export function TeamSubmissionResult({ payment }: { payment: Payment }) {
  const { authorizationHeader } = useAuth();
  const reference = teamPaymentReferenceSchema.safeParse(payment.reference);
  const teamId = reference.success ? reference.data.split(":")[0] : "";
  const isTeamPayment = payment.source === "TEAM_REGISTRATION" && !!teamId;
  const submission = useQuery({
    queryKey: ["team-submission", payment.id, payment.status],
    enabled:
      isTeamPayment && payment.status !== "UNPAID" && !!authorizationHeader,
    queryFn: async () => {
      const result = await apiClient.POST(
        "/api/v1/lobbies/{id}/submission/refresh",
        {
          params: { path: { id: teamId } },
          headers: { Authorization: authorizationHeader },
        }
      );
      if (result.error || !result.data)
        throw new Error("Unable to verify team submission");
      return result.data;
    },
    retry: 2,
  });
  if (!isTeamPayment || payment.status === "UNPAID") return null;
  const team = submission.data;
  const submitted =
    team && ["LOCKED", "ACCEPTED", "REJECTED"].includes(team.status);
  return (
    <Alert aria-live="polite">
      <AlertTitle>
        {submitted
          ? "Team application submitted"
          : team?.status === "FORMING"
            ? "Roster unlocked"
            : "Verifying team submission"}
      </AlertTitle>
      <AlertDescription>
        {submitted
          ? "Your application is with the organizer. Open your team to see its approval status."
          : team?.status === "FORMING"
            ? "Payment was unsuccessful or expired. Your team was not submitted. Return to the lobby to edit your roster or try again."
            : "Your payment result is saved. We are checking the team application; do not create another payment."}
        {submission.isError && (
          <Button
            variant="outline"
            onClick={() => void submission.refetch()}
            disabled={submission.isFetching}
          >
            Retry verification
          </Button>
        )}
        {team && (
          <Link className="underline" href={`/lobbies/${team.inviteCode}`}>
            Return to team lobby
          </Link>
        )}
      </AlertDescription>
    </Alert>
  );
}
