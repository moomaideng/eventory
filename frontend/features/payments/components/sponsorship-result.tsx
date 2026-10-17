"use client";

import Link from "next/link";
import { useEffect } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
import { paymentIdSchema } from "@/features/payments/schemas";
import type { Payment } from "@/features/payments/utils";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

export function SponsorshipResult({
  payment,
}: {
  payment: Pick<Payment, "id" | "source" | "reference" | "status">;
}) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();
  const pledgeId = paymentIdSchema.safeParse(payment.reference);
  const enabled =
    payment.source === "SPONSORSHIP" &&
    pledgeId.success &&
    Boolean(authorizationHeader);
  const query = useQuery({
    queryKey: ["sponsorship-result", payment.id, payment.status],
    enabled,
    queryFn: async () => {
      const result = await apiClient.POST(
        "/api/v1/tournaments/sponsorships/{id}/refresh",
        {
          params: { path: { id: payment.reference } },
          headers: { Authorization: authorizationHeader },
        }
      );
      if (result.error || !result.data)
        throw new Error("Unable to verify sponsorship");
      return result.data;
    },
    refetchInterval: (q) =>
      payment.status !== "UNPAID" && q.state.data?.status === "PENDING"
        ? 2000
        : false,
    retry: 2,
  });
  const pledge = query.data;
  useEffect(() => {
    if (!pledge || pledge.status === "PENDING") return;
    void queryClient.invalidateQueries({
      queryKey: ["get", "/api/v1/tournaments/sponsorships/mine"],
    });
    for (const path of [
      "/api/v1/tournaments/{tournamentId}/sponsorship-tiers",
      "/api/v1/tournaments/{tournamentId}/sponsors",
      "/api/v1/tournaments/{tournamentId}",
    ]) {
      void queryClient.invalidateQueries({
        queryKey: [
          "get",
          path,
          { params: { path: { tournamentId: pledge.tournamentId } } },
        ],
      });
    }
  }, [pledge, queryClient]);
  if (!enabled) return null;
  return (
    <Alert aria-live="polite">
      <AlertTitle>
        {pledge?.status === "CONFIRMED"
          ? "Sponsorship confirmed"
          : ["CANCELLED", "EXPIRED"].includes(pledge?.status ?? "")
            ? "Sponsorship not completed"
            : payment.status === "UNPAID"
              ? "Sponsorship slot reserved"
              : "Verifying sponsorship"}
      </AlertTitle>
      <AlertDescription>
        {pledge?.status === "CONFIRMED"
          ? "Your brand is now displayed on the tournament page."
          : payment.status === "UNPAID"
            ? "Complete payment before the deadline to confirm your package."
            : pledge?.status === "CANCELLED" || pledge?.status === "EXPIRED"
              ? "Your brand was not published. Return to the tournament to choose a package again."
              : "Your payment result is saved. We are checking the sponsorship; do not pay again."}
        {query.isError && (
          <Button
            variant="outline"
            onClick={() => void query.refetch()}
            disabled={query.isFetching}
          >
            Retry verification
          </Button>
        )}
        {pledge && (
          <Link
            href={`/sponsor/tournaments/${pledge.tournamentId}`}
            className="underline underline-offset-4"
          >
            Return to tournament
          </Link>
        )}
      </AlertDescription>
    </Alert>
  );
}
