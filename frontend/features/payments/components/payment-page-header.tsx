import { PageHeader } from "@/components/page-header";

export function PaymentPageHeader({
  title,
  description,
  invoiceId,
}: {
  title: string;
  description: string;
  invoiceId?: string;
}) {
  return (
    <PageHeader
      breadcrumbs={[
        {
          label: "My payments",
          href: "/payments",
        },
        ...(invoiceId
          ? [
              {
                label: "Invoice",
                href: `/payments/${invoiceId}`,
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
