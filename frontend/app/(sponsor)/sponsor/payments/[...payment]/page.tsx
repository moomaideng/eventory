import { notFound } from "next/navigation";
import { PaymentPageHeader } from "@/features/payments/components/payment-page-header";
import { Invoice } from "@/features/payments/components/invoice";
import { MockCheckout } from "@/features/payments/components/mock-checkout";
import { Receipt } from "@/features/payments/components/receipt";
import { sponsorPaymentRouteSchema } from "@/features/payments/schemas";

export const metadata = { title: "Sponsorship payment | Eventory" };

export default async function SponsorPaymentPage({
  params,
}: {
  params: Promise<{ payment: string[] }>;
}) {
  const parsed = sponsorPaymentRouteSchema.safeParse((await params).payment);
  if (!parsed.success) notFound();
  const [paymentId, view] = parsed.data;
  const title =
    view === "checkout"
      ? "Checkout"
      : view === "receipt"
        ? "Receipt"
        : "Invoice";
  return (
    <div className="sponsorship-surface mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 px-4 py-10 sm:px-8">
      <PaymentPageHeader
        title={title}
        description={
          view === "receipt"
            ? "View your payment confirmation or download the receipt as a PDF."
            : "Review your sponsorship payment and its deadline."
        }
        invoiceId={view ? paymentId : undefined}
        sponsorMode
      />
      {view === "checkout" ? (
        <MockCheckout paymentId={paymentId} sponsorMode />
      ) : view === "receipt" ? (
        <Receipt paymentId={paymentId} sponsorMode />
      ) : (
        <Invoice paymentId={paymentId} sponsorMode />
      )}
    </div>
  );
}
