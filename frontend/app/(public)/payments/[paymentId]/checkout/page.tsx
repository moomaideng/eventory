import { notFound } from "next/navigation";
import { CreditCard } from "lucide-react";
import { PaymentPageHeader } from "@/features/payments/components/payment-page-header";
import { MockCheckout } from "@/features/payments/components/mock-checkout";
import { paymentIdSchema } from "@/features/payments/schemas";

export const metadata = { title: "Checkout | Eventory" };

export default async function CheckoutPage({
  params,
}: {
  params: Promise<{ paymentId: string }>;
}) {
  const { paymentId } = await params;
  const parsed = paymentIdSchema.safeParse(paymentId);
  if (!parsed.success) notFound();
  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col gap-8 px-4 py-10">
      <PaymentPageHeader
        title="Checkout"
        description="Review your invoice and complete payment before the deadline."
        icon={CreditCard}
        invoiceId={parsed.data}
      />
      <MockCheckout paymentId={parsed.data} />
    </div>
  );
}
