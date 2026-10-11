import { formatPaymentDate, type Payment } from "@/features/payments/utils";

export function InvoiceDetails({ invoice }: { invoice: Payment }) {
  return (
    <dl className="grid gap-3 text-sm">
      <div>
        <dt className="text-muted-foreground">Reference</dt>
        <dd className="break-all">
          {invoice.source} / {invoice.reference}
        </dd>
      </div>
      <div>
        <dt className="text-muted-foreground">Created (UTC)</dt>
        <dd>{formatPaymentDate(invoice.createdAt)}</dd>
      </div>
      <div>
        <dt className="text-muted-foreground">Payment deadline (UTC)</dt>
        <dd>{formatPaymentDate(invoice.expiresAt)}</dd>
      </div>
      {invoice.paidAt && (
        <div>
          <dt className="text-muted-foreground">Paid at (UTC)</dt>
          <dd>{formatPaymentDate(invoice.paidAt)}</dd>
        </div>
      )}
      <div>
        <dt className="text-muted-foreground">Invoice ID</dt>
        <dd className="font-mono text-xs break-all">{invoice.id}</dd>
      </div>
    </dl>
  );
}
