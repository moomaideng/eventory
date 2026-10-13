"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { teamPaymentReferenceSchema } from "@/features/payments/schemas";
import type { Payment } from "@/features/payments/utils";

export function SoloSubmissionResult({ payment }: { payment: Payment }) {
  const { authorizationHeader } = useAuth();
  const parsed = teamPaymentReferenceSchema.safeParse(payment.reference);
  const tournamentId = parsed.success ? parsed.data.split(":")[0] : "";
  const active = payment.source === "SOLO_REGISTRATION" && Boolean(tournamentId);
  const submission = useQuery({
    queryKey: ["solo-registration-result", payment.id, payment.status],
    enabled: active && payment.status !== "UNPAID" && Boolean(authorizationHeader),
    retry: 2,
    queryFn: async () => {
      const result = await apiClient.POST(
        "/api/v1/tournaments/{tournamentId}/solo-registration/refresh",
        { params: { path: { tournamentId } }, headers: { Authorization: authorizationHeader } }
      );
      if (result.error || !result.data) throw new Error("Unable to verify solo registration");
      return result.data;
    },
  });
  if (!active || payment.status === "UNPAID") return null;
  const status = submission.data?.status;
  return (
    <Alert aria-live="polite">
      <AlertTitle>
        {status === "REGISTERED" ? "Registration complete" : status === "PAYMENT_FAILED" ? "Registration not submitted" : "Verifying registration"}
      </AlertTitle>
      <AlertDescription className="flex flex-col gap-2">
        <p>
          {status === "REGISTERED"
            ? "Your payment and individual registration are confirmed."
            : status === "PAYMENT_FAILED"
              ? "Payment was unsuccessful or expired. Your seat is available again; you can retry registration."
              : "Your payment result is saved. We are checking the registration status."}
        </p>
        {submission.isError && (
          <Button variant="outline" onClick={() => void submission.refetch()} disabled={submission.isFetching}>
            Retry verification
          </Button>
        )}
        <Link className="underline" href={`/tournaments/${tournamentId}`}>
          Return to tournament
        </Link>
      </AlertDescription>
    </Alert>
  );
}
