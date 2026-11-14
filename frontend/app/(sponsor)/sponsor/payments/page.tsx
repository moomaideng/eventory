import { PageHeader } from "@/components/page-header";
import { PaymentHistory } from "@/features/payments/components/payment-history";

export const metadata = { title: "My payments | Eventory" };

export default function SponsorPaymentsPage() {
  return (
    <div className="sponsorship-surface mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PageHeader
        workspace="sponsor"
        title="My payments"
        description="Your invoices and receipts across Eventory. Open an invoice to pay or view its receipt."
      />
      <PaymentHistory sponsorMode />
    </div>
  );
}
