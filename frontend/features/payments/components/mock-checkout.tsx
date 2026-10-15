"use client";

import Link from "next/link";
import { TeamSubmissionResult } from "@/features/payments/components/team-submission-result";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { Info } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import { PaymentStatus } from "@/features/payments/components/payment-display";
import { useInvoice } from "@/features/payments/hooks/use-invoice";
import { formatAmount, formatPaymentDate } from "@/features/payments/utils";

export function MockCheckout({ paymentId }: { paymentId: string }) {
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();
  const [lastResult, setLastResult] = useState<"FAILED" | "CANCELLED" | null>(
    null
  );
  const options = $api.queryOptions("get", "/api/v1/payments/{id}", {
    params: { path: { id: paymentId } },
    headers: { Authorization: authorizationHeader },
  });
  const paymentQuery = useInvoice(paymentId);
  const resultMutation = $api.useMutation(
    "post",
    "/api/v1/payments/{id}/mock-result",
    {
      onSuccess: (payment, variables) => {
        const result = variables.body.result;
        setLastResult(result === "SUCCEEDED" ? null : result);
        queryClient.setQueryData(options.queryKey, payment);
        void queryClient.invalidateQueries({
          queryKey: ["get", "/api/v1/payments"],
        });
      },
      onError: () => {
        void paymentQuery.refetch();
      },
    }
  );
  const complete = (result: "SUCCEEDED" | "FAILED" | "CANCELLED") => {
    resultMutation.mutate({
      params: { path: { id: paymentId } },
      headers: { Authorization: authorizationHeader },
      body: { result },
    });
  };
  if (paymentQuery.isPending)
    return <Skeleton className="h-80 w-full" aria-label="Loading payment" />;
  if (paymentQuery.isError)
    return (
      <Alert variant="destructive">
        <AlertTitle>Payment unavailable</AlertTitle>
        <AlertDescription>
          This payment may not belong to you, or the service could not load it.
          <Button variant="outline" onClick={() => void paymentQuery.refetch()}>
            Try again
          </Button>
          <Link href="/payments" className="underline">
            Back to my payments
          </Link>
        </AlertDescription>
      </Alert>
    );
  const payment = paymentQuery.data;
  return (
    <div className="flex flex-col gap-6">
      <TeamSubmissionResult payment={payment} />
      <Alert>
        <Info />
        <AlertTitle>Demo checkout</AlertTitle>
        <AlertDescription>
          This payment is simulated. No real money is transferred. Choose a
          result below to try the payment flow.
        </AlertDescription>
      </Alert>
      <Card>
        <CardHeader>
          <CardTitle>Complete payment</CardTitle>
          <CardDescription className="break-words">
            {payment.description}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-6">
          <div className="flex flex-col items-center gap-3 py-4">
            <PaymentStatus status={payment.status} />
            <p className="w-full text-center text-2xl font-semibold break-words tabular-nums sm:text-4xl">
              {formatAmount(payment)}
            </p>
            <p className="text-muted-foreground text-center text-sm">
              Payment deadline: {formatPaymentDate(payment.expiresAt)} (UTC)
            </p>
          </div>
          {resultMutation.isError && (
            <Alert variant="destructive">
              <AlertTitle>Payment could not be completed</AlertTitle>
              <AlertDescription>
                Check the current status before trying again. Paid or expired
                invoices cannot accept another payment.
              </AlertDescription>
            </Alert>
          )}
          {payment.status === "UNPAID" && lastResult && (
            <Alert>
              <AlertTitle>
                {lastResult === "FAILED"
                  ? "Payment failed"
                  : "Payment cancelled"}
              </AlertTitle>
              <AlertDescription>
                No payment was made. You can retry this invoice before its
                deadline.
              </AlertDescription>
            </Alert>
          )}
          {payment.status !== "UNPAID" && (
            <Alert>
              <AlertTitle>
                {payment.status === "PAID"
                  ? "Payment successful"
                  : payment.status === "CANCELLED"
                    ? lastResult === "FAILED"
                      ? "Payment failed"
                      : "Payment cancelled"
                    : "Invoice expired"}
              </AlertTitle>
              <AlertDescription>
                {payment.status === "PAID"
                  ? "Your payment has been recorded. View your receipt below."
                  : "This invoice is closed and can no longer be paid. No payment was made."}
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
        <CardFooter className="flex flex-col items-stretch gap-3">
          {payment.status === "UNPAID" ? (
            <>
              <Button
                size="lg"
                disabled={resultMutation.isPending}
                onClick={() => complete("SUCCEEDED")}
              >
                {resultMutation.isPending && (
                  <Spinner data-icon="inline-start" />
                )}
                Simulate success
              </Button>
              <div className="grid grid-cols-2 gap-3">
                <Button
                  variant="outline"
                  disabled={resultMutation.isPending}
                  onClick={() => complete("FAILED")}
                >
                  Simulate failure
                </Button>
                <Button
                  variant="ghost"
                  disabled={resultMutation.isPending}
                  onClick={() => complete("CANCELLED")}
                >
                  Cancel payment
                </Button>
              </div>
            </>
          ) : (
            <Link
              href={
                payment.status === "PAID"
                  ? `/payments/${payment.id}/receipt`
                  : payment.invoicePath
              }
              className={buttonVariants({ variant: "outline" })}
            >
              {payment.status === "PAID" ? "View receipt" : "View invoice"}
            </Link>
          )}
        </CardFooter>
      </Card>
      <Link href={payment.invoicePath} className="text-sm underline">
        Back to invoice
      </Link>
    </div>
  );
}
