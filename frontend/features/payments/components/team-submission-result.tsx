"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { RegistrationPaymentButton } from "@/features/payments/components/registration-payment-button";
import { teamPaymentReferenceSchema } from "@/features/payments/schemas";
import type { Payment } from "@/features/payments/utils";

export function TeamSubmissionResult({ payment }: { payment: Payment }) {
  const { authorizationHeader } = useAuth();
  const reference = teamPaymentReferenceSchema.safeParse(payment.reference);
  const [tournamentId = "", teamId = ""] = reference.success
    ? reference.data.split(":")
    : [];
  const isTeamPayment =
    payment.source === "TEAM_REGISTRATION" && !!tournamentId && !!teamId;
  const submission = useQuery({
    queryKey: ["team-submission", payment.id, payment.status],
    enabled:
      isTeamPayment && payment.status !== "UNPAID" && payment.status !== "REFUNDED" && !!authorizationHeader,
    queryFn: async () => {
      const result = await apiClient.POST(
        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/submission/refresh",
        {
          params: { path: { tournamentId, teamId } },
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
  const isSolo = Boolean(team) && !team?.inviteCode;
  return (
    <Alert aria-live="polite">
      <AlertTitle>
        {team?.status === "REJECTED"
          ? "Application rejected"
          : submitted
          ? isSolo
            ? "Registration submitted"
            : "Team application submitted"
          : team?.status === "PAYMENT_PENDING"
            ? "Payment required"
          : team?.status === "FORMING"
            ? isSolo
              ? "Registration not completed"
              : "Roster unlocked"
            : "Verifying team submission"}
      </AlertTitle>
      <AlertDescription>
        {team?.status === "REJECTED"
          ? "Your organizer rejected this registration. If you paid a fee, its refund is being processed."
          : submitted
          ? isSolo
            ? "Your registration is with the organizer for review."
            : "Your application is with the organizer. Open your team to see its approval status."
          : team?.status === "PAYMENT_PENDING"
            ? "Your payment was unsuccessful or expired. Your roster and answers are saved. Try payment again to complete registration."
          : team?.status === "FORMING"
            ? isSolo
              ? "Payment was unsuccessful or expired. Your registration was not submitted. Go back to the tournament to try again."
              : "Payment was unsuccessful or expired. Your team was not submitted. Return to the lobby to edit your roster or try again."
            : "Your payment result is saved. We are checking the team application; do not create another payment."}
        {team && ["CANCELLED", "EXPIRED"].includes(payment.status) && ["FORMING", "PAYMENT_PENDING"].includes(team.status) && (
          <RegistrationPaymentButton tournamentId={team.tournament.id} teamId={team.id}>
            Try payment again
          </RegistrationPaymentButton>
        )}
        {submission.isError && (
          <Button
            variant="outline"
            onClick={() => void submission.refetch()}
            disabled={submission.isFetching}
          >
            Retry verification
          </Button>
        )}
        {team?.inviteCode ? (
          <Link
            className="mt-2 block underline"
            href={`/tournaments/${team.tournament.id}/teams/${team.id}`}
          >
            Return to team lobby
          </Link>
        ) : team ? (
          <Link
            className="mt-2 block underline"
            href={`/tournaments/${team.tournament.id}`}
          >
            Back to the tournament
          </Link>
        ) : null}
      </AlertDescription>
    </Alert>
  );
}
