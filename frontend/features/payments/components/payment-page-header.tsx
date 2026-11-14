import { PageHeader } from "@/components/page-header";

export function PaymentPageHeader({
  title,
  description,
  invoiceId,
  sponsorMode = false,
}: {
  title: string;
  description: string;
  invoiceId?: string;
  sponsorMode?: boolean;
}) {
  return (
    <PageHeader
      workspace={sponsorMode ? "sponsor" : undefined}
      breadcrumbs={[
        {
          label: "My payments",
          href: sponsorMode ? "/sponsor/payments" : "/payments",
        },
        ...(invoiceId
          ? [
              {
                label: "Invoice",
                href: `${sponsorMode ? "/sponsor" : ""}/payments/${invoiceId}`,
              },
            ]
          : []),
        { label: title },
      ]}
      title={title}
      description={description}
    />
  );
}
