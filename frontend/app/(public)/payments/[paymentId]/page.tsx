import { notFound } from "next/navigation";
import { FileText } from "lucide-react";
import { PaymentPageHeader } from "@/features/payments/components/payment-page-header";
import { Invoice } from "@/features/payments/components/invoice";
import { paymentIdSchema } from "@/features/payments/schemas";

export const metadata = { title: "Invoice | Eventory" };

export default async function InvoicePage({
  params,
}: {
  params: Promise<{ paymentId: string }>;
}) {
  const { paymentId } = await params;
  const parsed = paymentIdSchema.safeParse(paymentId);
  if (!parsed.success) notFound();
  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-8 px-4 py-10">
      <PaymentPageHeader
        title="Invoice"
        description="Review the payment request, amount and deadline before checkout."
        icon={FileText}
      />
      <Invoice paymentId={parsed.data} />
    </div>
  );
}
