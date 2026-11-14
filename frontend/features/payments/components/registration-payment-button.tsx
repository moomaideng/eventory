"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
import { extractProblemMessage } from "@/lib/schemas";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";

export function RegistrationPaymentButton({
  tournamentId,
  teamId,
  children = "Continue to payment",
}: {
  tournamentId: string;
  teamId: string;
  children?: React.ReactNode;
}) {
  const { authorizationHeader } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");

  async function continuePayment() {
    if (!authorizationHeader || pending) return;
    setPending(true);
    setError("");
    try {
      // Resolve the active attempt on Tournament before opening checkout.
      const result = await apiClient.POST(
        "/api/v1/tournaments/{tournamentId}/teams/{teamId}/lock",
        {
          params: { path: { tournamentId, teamId } },
          headers: { Authorization: authorizationHeader },
        }
      );
      if (result.error || !result.data) {
        setError(extractProblemMessage(result.error, "Unable to continue registration. Please retry."));
        return;
      }
      const team = result.data;
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["get", "/api/v1/tournaments/{tournamentId}/my-team"] }),
        queryClient.invalidateQueries({
          queryKey: ["get", "/api/v1/tournaments/{tournamentId}/teams/{teamId}"],
        }),
        queryClient.invalidateQueries({ queryKey: ["team-submission"] }),
      ]);
      if (team.status === "PAYMENT_PENDING" && team.paymentId) {
        router.push(`/payments/${team.paymentId}/checkout`);
      } else if (["LOCKED", "ACCEPTED", "REJECTED"].includes(team.status)) {
        router.push(
          team.inviteCode
            ? `/tournaments/${team.tournament.id}/teams/${team.id}`
            : `/tournaments/${team.tournament.id}`
        );
      } else {
        setError("Payment could not be started. Please retry.");
      }
    } catch {
      setError("Connection error. Please check your registration before retrying.");
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="flex flex-col items-start gap-2">
      {error && <Alert variant="destructive"><AlertDescription>{error}</AlertDescription></Alert>}
      <Button disabled={pending || !authorizationHeader} onClick={() => void continuePayment()}>
        {pending && <Spinner data-icon="inline-start" />}
        {children}
      </Button>
    </div>
  );
}
