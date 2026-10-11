import { notFound } from "next/navigation";
import { ReceiptText } from "lucide-react";
import { PaymentPageHeader } from "@/features/payments/components/payment-page-header";
import { Receipt } from "@/features/payments/components/receipt";
import { paymentIdSchema } from "@/features/payments/schemas";

export const metadata = { title: "Receipt | Eventory" };

export default async function ReceiptPage({
  params,
}: {
  params: Promise<{ paymentId: string }>;
}) {
  const { paymentId } = await params;
  const parsed = paymentIdSchema.safeParse(paymentId);
  if (!parsed.success) notFound();
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-8 px-4 py-10">
      <PaymentPageHeader
        title="Receipt"
        description="View your payment confirmation or download the receipt as a PDF."
        icon={ReceiptText}
        invoiceId={parsed.data}
      />
      <Receipt paymentId={parsed.data} />
    </div>
  );
}
