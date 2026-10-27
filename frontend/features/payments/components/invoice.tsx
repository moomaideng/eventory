"use client";

import Link from "next/link";
import { TeamSubmissionResult } from "@/features/payments/components/team-submission-result";
import { SponsorshipResult } from "@/features/payments/components/sponsorship-result";
import { useInvoice } from "@/features/payments/hooks/use-invoice";
import { formatAmount } from "@/features/payments/utils";
import { InvoiceDetails } from "@/features/payments/components/invoice-details";
import { PaymentStatus } from "@/features/payments/components/payment-display";
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

export function Invoice({
  paymentId,
  sponsorMode = false,
}: {
  paymentId: string;
  sponsorMode?: boolean;
}) {
  const basePath = sponsorMode ? "/sponsor/payments" : "/payments";
  const query = useInvoice(paymentId);
  if (query.isPending)
    return <Skeleton className="h-80 w-full" aria-label="Loading invoice" />;
  if (query.isError)
    return (
      <Alert variant="destructive">
        <AlertTitle>Invoice unavailable</AlertTitle>
        <AlertDescription>
          This invoice may not belong to you, or could not be loaded.
          <Button variant="outline" onClick={() => void query.refetch()}>
            Try again
          </Button>
          <Link href={basePath} className="underline">
            Back to my payments
          </Link>
        </AlertDescription>
      </Alert>
    );
  const invoice = query.data;
  return (
    <div className="flex flex-col gap-6">
      <TeamSubmissionResult payment={invoice} />
      <SponsorshipResult payment={invoice} />
      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <Card>
          <CardHeader>
            <CardTitle>Payment request</CardTitle>
            <CardDescription className="break-words">
              {invoice.description}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <InvoiceDetails invoice={invoice} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>
              {invoice.status === "REFUNDED"
                ? "Amount refunded"
                : invoice.status === "PAID"
                  ? "Amount paid"
                  : "Amount due"}
            </CardTitle>
            <CardDescription>Set by the requesting service.</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-6">
            <div className="flex flex-col items-start gap-3">
              <PaymentStatus status={invoice.status} />
              <p className="w-full text-2xl font-semibold break-words tabular-nums">
                {formatAmount(invoice)}
              </p>
            </div>
            <Alert>
              <AlertTitle>
                {invoice.status === "REFUNDED"
                  ? "Payment refunded"
                  : invoice.status === "PAID"
                  ? "Invoice paid"
                  : invoice.status === "CANCELLED"
                    ? "Payment attempt closed"
                    : invoice.status === "EXPIRED"
                      ? "Payment deadline has passed"
                      : "Payment is due"}
              </AlertTitle>
              <AlertDescription>
                {invoice.status === "REFUNDED"
                  ? "Your organizer rejected the registration. The full amount has been refunded; your original payment receipt is still available."
                  : invoice.status === "PAID"
                  ? "Your receipt is available below."
                  : invoice.status === "CANCELLED"
                    ? "No payment was made. Return to the tournament or team lobby to start again."
                    : invoice.status === "EXPIRED"
                      ? "This invoice can no longer be paid. Contact the requesting service for a new payment request."
                      : "Complete payment before the deadline. Failed or cancelled payments can be retried."}
              </AlertDescription>
            </Alert>
          </CardContent>
          <CardFooter className="flex flex-wrap gap-3">
            {invoice.status === "UNPAID" && (
              <Link
                href={`${basePath}/${invoice.id}/checkout`}
                className={buttonVariants()}
              >
                Proceed to checkout
              </Link>
            )}
            {(invoice.status === "PAID" || invoice.status === "REFUNDED") && (
              <Link
                href={`${basePath}/${invoice.id}/receipt`}
                className={buttonVariants()}
              >
                View receipt
              </Link>
            )}
            <Link
              href={basePath}
              className={buttonVariants({ variant: "outline" })}
            >
              My payments
            </Link>
          </CardFooter>
        </Card>
      </div>
    </div>
  );
}
