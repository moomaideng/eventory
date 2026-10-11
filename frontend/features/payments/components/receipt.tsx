"use client";

import { useState } from "react";
import Link from "next/link";
import { Download, BadgeCheck } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api, apiClient } from "@/lib/api/client";
import { formatAmount, formatPaymentDate } from "@/features/payments/utils";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardHeader,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import { Spinner } from "@/components/ui/spinner";

export function Receipt({ paymentId }: { paymentId: string }) {
  const { authorizationHeader } = useAuth();
  const [downloading, setDownloading] = useState(false);
  const [downloadError, setDownloadError] = useState(false);
  const query = $api.useQuery(
    "get",
    "/api/v1/payments/{id}/receipt",
    {
      params: { path: { id: paymentId } },
      headers: { Authorization: authorizationHeader },
    },
    { enabled: !!authorizationHeader }
  );

  async function download() {
    if (!query.data || downloading) return;
    setDownloading(true);
    setDownloadError(false);
    try {
      const { data, response } = await apiClient.GET(
        "/api/v1/payments/{id}/receipt.pdf",
        {
          params: { path: { id: paymentId } },
          headers: { Authorization: authorizationHeader },
          parseAs: "blob",
        }
      );
      if (!response.ok || !data || data.type !== "application/pdf")
        throw new Error("Receipt download failed");
      const url = URL.createObjectURL(data);
      const link = document.createElement("a");
      link.href = url;
      link.download = `receipt-${query.data.id}.pdf`;
      document.body.append(link);
      link.click();
      link.remove();
      // Give the browser time to start reading the blob before releasing it.
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    } catch {
      setDownloadError(true);
    } finally {
      setDownloading(false);
    }
  }
  if (query.isPending)
    return <Skeleton className="h-80 w-full" aria-label="Loading receipt" />;
  if (query.isError)
    return (
      <Alert variant="destructive">
        <AlertTitle>Receipt unavailable</AlertTitle>
        <AlertDescription>
          A receipt is only available after this invoice has been paid, and only
          to its owner.
          <Button variant="outline" onClick={() => void query.refetch()}>
            Try again
          </Button>
          <Link href={`/payments/${paymentId}`} className="underline">
            Back to invoice
          </Link>
        </AlertDescription>
      </Alert>
    );
  const receipt = query.data;
  return (
    <Card className="receipt-paper">
      <CardHeader>
        <div className="flex flex-col items-center gap-3 py-4 text-center">
          <BadgeCheck aria-hidden="true" className="text-primary size-14" />
          <h2 className="text-2xl font-semibold">Payment successful</h2>
          <p className="w-full text-2xl font-semibold break-words tabular-nums sm:text-4xl">
            {formatAmount(receipt)}
          </p>
          <p className="text-muted-foreground w-full text-sm break-words">
            {receipt.description}
          </p>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-6">
        <Separator />
        <div>
          <h3 className="font-medium">Receipt details</h3>
          <CardDescription>
            Paid at {formatPaymentDate(receipt.paidAt)} (UTC)
          </CardDescription>
        </div>
        <dl className="grid gap-5 text-sm sm:grid-cols-2">
          <div>
            <dt className="text-muted-foreground">Reference</dt>
            <dd className="break-all">
              {receipt.source} / {receipt.reference}
            </dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Receipt ID</dt>
            <dd className="font-mono text-xs break-all">{receipt.id}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Invoice ID</dt>
            <dd className="font-mono text-xs break-all">{receipt.invoiceId}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">Payer account</dt>
            <dd className="font-mono text-xs break-all">
              {receipt.payerAccountId}
            </dd>
          </div>
        </dl>
        {downloadError && (
          <Alert variant="destructive">
            <AlertTitle>Could not download the PDF</AlertTitle>
            <AlertDescription>Please try downloading again.</AlertDescription>
          </Alert>
        )}
      </CardContent>
      <CardFooter className="flex flex-wrap gap-3">
        <Button disabled={downloading} onClick={() => void download()}>
          {downloading ? (
            <Spinner data-icon="inline-start" />
          ) : (
            <Download data-icon="inline-start" />
          )}
          Download PDF
        </Button>
        <Link
          href={`/payments/${paymentId}`}
          className={buttonVariants({ variant: "outline" })}
        >
          View invoice
        </Link>
      </CardFooter>
    </Card>
  );
}
