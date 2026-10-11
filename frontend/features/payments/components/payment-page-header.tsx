import Link from "next/link";
import { ChevronRight, type LucideIcon } from "lucide-react";

export function PaymentPageHeader({
  title,
  description,
  icon: Icon,
  invoiceId,
}: {
  title: string;
  description: string;
  icon: LucideIcon;
  invoiceId?: string;
}) {
  return (
    <header className="flex flex-col gap-5">
      <nav aria-label="Breadcrumb">
        <ol className="text-muted-foreground flex flex-wrap items-center gap-2 text-sm">
          <li>
            <Link href="/payments" className="hover:underline">
              My payments
            </Link>
          </li>
          <li aria-hidden="true">
            <ChevronRight className="size-4" />
          </li>
          {invoiceId && (
            <>
              <li>
                <Link
                  href={`/payments/${invoiceId}`}
                  className="hover:underline"
                >
                  Invoice
                </Link>
              </li>
              <li aria-hidden="true">
                <ChevronRight className="size-4" />
              </li>
            </>
          )}
          <li>
            <span aria-current="page" className="text-foreground">
              {title}
            </span>
          </li>
        </ol>
      </nav>
      <div className="flex items-start gap-4">
        <Icon
          aria-hidden="true"
          className="text-primary mt-1 size-8 shrink-0"
        />
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-semibold tracking-tight">{title}</h1>
          <p className="text-muted-foreground text-sm">{description}</p>
        </div>
      </div>
    </header>
  );
}
