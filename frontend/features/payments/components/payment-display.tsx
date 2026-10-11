import { Badge } from "@/components/ui/badge";
import type { Payment } from "@/features/payments/utils";

const statusLabels: Record<Payment["status"], string> = {
  UNPAID: "Unpaid",
  PAID: "Paid",
  EXPIRED: "Expired",
};

export function PaymentStatus({ status }: { status: Payment["status"] }) {
  return (
    <Badge
      variant={
        status === "EXPIRED"
          ? "destructive"
          : status === "PAID"
            ? "default"
            : "secondary"
      }
    >
      {statusLabels[status]}
    </Badge>
  );
}
