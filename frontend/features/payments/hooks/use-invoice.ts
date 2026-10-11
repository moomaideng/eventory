"use client";

import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";

export function useInvoice(paymentId: string) {
  const { authorizationHeader } = useAuth();
  return $api.useQuery(
    "get",
    "/api/v1/payments/{id}",
    {
      params: { path: { id: paymentId } },
      headers: { Authorization: authorizationHeader },
    },
    {
      enabled: !!authorizationHeader,
      // Refresh the backend-derived status while the invoice can still expire.
      refetchInterval: (query) =>
        query.state.data?.status === "UNPAID" ? 5000 : false,
    }
  );
}
