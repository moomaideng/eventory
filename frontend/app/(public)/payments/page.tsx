import { PageHeader } from "@/components/page-header";
import { PaymentHistory } from "@/features/payments/components/payment-history";

export const metadata = { title: "My payments | Eventory" };

export default function PaymentsPage() {
  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 px-4 py-10">
      <PageHeader
        title="My payments"
        description="Track your invoices and open a row to view its details."
      />
      <PaymentHistory />
    </div>
  );
}
