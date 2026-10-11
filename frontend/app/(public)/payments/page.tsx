import { PaymentHistory } from "@/features/payments/components/payment-history";

export const metadata = { title: "My payments | Eventory" };

export default function PaymentsPage() {
  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-8 px-4 py-10">
      <header className="flex flex-col gap-2">
        <h1 className="text-3xl font-semibold tracking-tight">My payments</h1>
        <p className="text-muted-foreground text-sm">
          Track your invoices and open a row to view its details.
        </p>
      </header>
      <PaymentHistory />
    </div>
  );
}
