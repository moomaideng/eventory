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
      isTeamPayment && payment.status !== "UNPAID" && payment.status !== "REFUNDED" && !!authorizationHeader,
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
  if (payment.status === "REFUNDED")
    return (
      <Alert aria-live="polite">
        <AlertTitle>Application rejected</AlertTitle>
        <AlertDescription>Your full registration fee has been refunded.</AlertDescription>
      </Alert>
    );
  const team = submission.data;
  const submitted =
    team && ["LOCKED", "ACCEPTED", "REJECTED"].includes(team.status);
  return (
    <Alert aria-live="polite">
      <AlertTitle>
        {team?.status === "REJECTED"
          ? "Application rejected"
          : submitted
          ? "Team application submitted"
          : team?.status === "FORMING"
            ? "Roster unlocked"
            : "Verifying team submission"}
      </AlertTitle>
      <AlertDescription>
        {team?.status === "REJECTED"
          ? "Your organizer rejected this registration. If you paid a fee, its refund is being processed."
          : submitted
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
        {team && team.inviteCode && (
          <Link className="underline" href={`/lobbies/${team.inviteCode}`}>
            Return to team lobby
          </Link>
        )}
      </AlertDescription>
    </Alert>
  );
}
