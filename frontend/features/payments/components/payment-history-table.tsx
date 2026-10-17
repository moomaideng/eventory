import Link from "next/link";
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
} from "@/components/ui/table";
import { PaymentStatus } from "@/features/payments/components/payment-display";
import {
  formatAmount,
  formatPaymentDate,
  type Payment,
} from "@/features/payments/utils";

export function PaymentHistoryTable({
  payments,
  sponsorMode = false,
}: {
  payments: Payment[];
  sponsorMode?: boolean;
}) {
  return (
    <Table>
      <TableCaption>
        Amounts are in THB. Dates and filters use UTC.
      </TableCaption>
      <TableHeader>
        <TableRow>
          <TableHead>Invoice</TableHead>
          <TableHead className="hidden sm:table-cell">Created</TableHead>
          <TableHead className="hidden sm:table-cell">Amount</TableHead>
          <TableHead className="hidden sm:table-cell">Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {payments.map((payment) => (
          <TableRow
            key={payment.id}
            className="focus-within:bg-muted/50 relative cursor-pointer"
          >
            <TableCell className="whitespace-normal">
              <p className="max-w-48 break-words sm:max-w-64">
                <Link
                  href={`${sponsorMode ? "/sponsor" : ""}/payments/${payment.id}`}
                  aria-label={`View invoice: ${payment.description}`}
                  className="focus-visible:after:outline-ring after:absolute after:inset-0 focus-visible:outline-none focus-visible:after:outline-2 focus-visible:after:-outline-offset-2"
                >
                  {payment.description}
                </Link>
              </p>
              <p className="text-muted-foreground max-w-48 truncate text-xs sm:max-w-64">
                {payment.source} / {payment.reference}
              </p>
              <div className="mt-2 flex flex-col gap-2 sm:hidden">
                <p className="text-muted-foreground text-xs">
                  {formatPaymentDate(payment.createdAt)}
                </p>
                <div className="flex flex-wrap items-center gap-2">
                  <span className="tabular-nums">{formatAmount(payment)}</span>
                  <PaymentStatus status={payment.status} />
                </div>
              </div>
            </TableCell>
            <TableCell className="hidden sm:table-cell">
              {formatPaymentDate(payment.createdAt)}
            </TableCell>
            <TableCell className="hidden tabular-nums sm:table-cell">
              {formatAmount(payment)}
            </TableCell>
            <TableCell className="hidden sm:table-cell">
              <PaymentStatus status={payment.status} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
