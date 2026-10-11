"use client";

import { useState, type FormEvent } from "react";
import { ReceiptText } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { historyDatesSchema } from "@/features/payments/schemas";
import { PaymentHistoryTable } from "@/features/payments/components/payment-history-table";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card";
import {
  Empty,
  EmptyHeader,
  EmptyTitle,
  EmptyDescription,
  EmptyMedia,
} from "@/components/ui/empty";
import {
  Field,
  FieldGroup,
  FieldLabel,
  FieldError,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";

const limit = 20;

export function PaymentHistory() {
  const { authorizationHeader } = useAuth();
  const [dates, setDates] = useState({ from: "", to: "" });
  const [filter, setFilter] = useState({ from: "", to: "", page: 1 });
  const [errors, setErrors] = useState<{ from?: string; to?: string }>({});
  const history = $api.useQuery(
    "get",
    "/api/v1/payments",
    {
      params: {
        query: {
          from: filter.from || undefined,
          to: filter.to || undefined,
          page: filter.page,
          limit,
        },
      },
      headers: { Authorization: authorizationHeader },
    },
    {
      enabled: !!authorizationHeader,
      refetchInterval: (query) =>
        query.state.data?.items?.some((invoice) => invoice.status === "UNPAID")
          ? 5000
          : false,
    }
  );
  function applyDates(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const parsed = historyDatesSchema.safeParse(dates);
    if (!parsed.success) {
      setErrors(
        Object.fromEntries(
          parsed.error.issues.map((issue) => [issue.path[0], issue.message])
        )
      );
      return;
    }
    setErrors({});
    setFilter({ ...parsed.data, page: 1 });
  }
  function clearDates() {
    setDates({ from: "", to: "" });
    setFilter({ from: "", to: "", page: 1 });
    setErrors({});
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>Invoice history</CardTitle>
        <CardDescription>
          View your invoices, payment status and receipts.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-6">
        <form onSubmit={applyDates}>
          <FieldGroup className="flex flex-col gap-4 sm:flex-row sm:items-end">
            <Field data-invalid={!!errors.from}>
              <FieldLabel htmlFor="payment-from">From (UTC)</FieldLabel>
              <Input
                id="payment-from"
                type="date"
                value={dates.from}
                aria-invalid={!!errors.from}
                aria-describedby={
                  errors.from ? "payment-from-error" : undefined
                }
                onChange={(event) =>
                  setDates({ ...dates, from: event.target.value })
                }
              />
              {errors.from && (
                <FieldError id="payment-from-error">{errors.from}</FieldError>
              )}
            </Field>
            <Field data-invalid={!!errors.to}>
              <FieldLabel htmlFor="payment-to">To (UTC)</FieldLabel>
              <Input
                id="payment-to"
                type="date"
                value={dates.to}
                aria-invalid={!!errors.to}
                aria-describedby={errors.to ? "payment-to-error" : undefined}
                onChange={(event) =>
                  setDates({ ...dates, to: event.target.value })
                }
              />
              {errors.to && (
                <FieldError id="payment-to-error">{errors.to}</FieldError>
              )}
            </Field>
            <div className="flex shrink-0 gap-2">
              <Button type="submit">Apply</Button>
              <Button type="button" variant="outline" onClick={clearDates}>
                Clear
              </Button>
            </div>
          </FieldGroup>
        </form>
        {history.isPending ? (
          <Skeleton
            className="h-64 w-full"
            aria-label="Loading payment history"
          />
        ) : history.isError ? (
          <Alert variant="destructive">
            <AlertTitle>Could not load your payments</AlertTitle>
            <AlertDescription>
              <Button variant="outline" onClick={() => void history.refetch()}>
                Try again
              </Button>
            </AlertDescription>
          </Alert>
        ) : !history.data.items?.length ? (
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ReceiptText />
              </EmptyMedia>
              <EmptyTitle>No invoices found</EmptyTitle>
              <EmptyDescription>
                {filter.from || filter.to
                  ? "Try a different date range."
                  : "Invoices will appear here when a service creates one for you."}
              </EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <PaymentHistoryTable payments={history.data.items} />
        )}
      </CardContent>
      <CardFooter className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-muted-foreground text-sm" aria-live="polite">
          {history.data
            ? `${history.data.total} invoices · Page ${filter.page}`
            : "Loading history…"}
        </p>
        <div className="flex gap-2">
          <Button
            variant="outline"
            disabled={filter.page === 1 || history.isFetching}
            onClick={() => setFilter({ ...filter, page: filter.page - 1 })}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            disabled={
              !history.data ||
              filter.page * limit >= history.data.total ||
              history.isFetching
            }
            onClick={() => setFilter({ ...filter, page: filter.page + 1 })}
          >
            Next
          </Button>
        </div>
      </CardFooter>
    </Card>
  );
}
