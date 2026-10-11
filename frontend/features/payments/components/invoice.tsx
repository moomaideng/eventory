"use client";

import Link from "next/link";
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

export function Invoice({ paymentId }: { paymentId: string }) {
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
          <Link href="/payments" className="underline">
            Back to my payments
          </Link>
        </AlertDescription>
      </Alert>
    );
  const invoice = query.data;
  return (
    <div className="flex flex-col gap-6">
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
              {invoice.status === "PAID" ? "Amount paid" : "Amount due"}
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
                {invoice.status === "PAID"
                  ? "Paid through mock checkout"
                  : invoice.status === "EXPIRED"
                    ? "Payment deadline has passed"
                    : "Payment is due"}
              </AlertTitle>
              <AlertDescription>
                {invoice.status === "PAID"
                  ? "Your mock receipt is available below. No real money was transferred."
                  : invoice.status === "EXPIRED"
                    ? "This invoice can no longer be paid. Contact the requesting service for a new payment request."
                    : "Complete mock checkout before the deadline. Failed or cancelled checkouts leave this invoice unpaid and can be retried."}
              </AlertDescription>
            </Alert>
          </CardContent>
          <CardFooter className="flex flex-wrap gap-3">
            {invoice.status === "UNPAID" && (
              <Link href={invoice.checkoutPath} className={buttonVariants()}>
                Go to mock checkout
              </Link>
            )}
            {invoice.status === "PAID" && (
              <Link
                href={`/payments/${invoice.id}/receipt`}
                className={buttonVariants()}
              >
                View receipt
              </Link>
            )}
            <Link
              href="/payments"
              className={buttonVariants({ variant: "outline" })}
            >
              My payments
            </Link>
          </CardFooter>
        </Card>
      </div>
      <p className="text-muted-foreground text-sm">
        Mock payment only. Not connected to tournament registration or
        sponsorship yet.
      </p>
    </div>
  );
}
