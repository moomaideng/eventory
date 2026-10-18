"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiBaseUrl } from "@/lib/api/client";
import { extractProblemMessage } from "@/lib/schemas";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import {
  sponsorshipBrandSchema,
  sponsorshipCheckoutSchema,
} from "@/features/sponsorship/schemas";
import {
  checkoutErrorFallback,
  type SponsorshipTier,
} from "@/features/sponsorship/utils";
import { formatMoney } from "@/features/tournaments/utils";

export function SponsorshipCheckoutDialog({
  tournamentId,
  tier,
  currency,
  sponsorName,
  onClose,
}: {
  tournamentId: string;
  tier: SponsorshipTier;
  currency: string;
  sponsorName: string;
  onClose: () => void;
}) {
  const { authorizationHeader } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [websiteUrl, setWebsiteUrl] = useState("");
  const [logo, setLogo] = useState<File>();
  const [preview, setPreview] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitError, setSubmitError] = useState("");
  const [busy, setBusy] = useState(false);
  const [requestId, setRequestId] = useState(() => crypto.randomUUID());

  useEffect(() => {
    return () => {
      if (preview) URL.revokeObjectURL(preview);
    };
  }, [preview]);

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy || !authorizationHeader) return;
    const parsed = sponsorshipBrandSchema.safeParse({ websiteUrl, logo });
    if (!parsed.success) {
      setErrors(
        Object.fromEntries(
          parsed.error.issues.map((issue) => [
            String(issue.path[0]),
            issue.message,
          ])
        )
      );
      return;
    }
    setErrors({});
    setSubmitError("");
    setBusy(true);
    const body = new FormData();
    body.set("requestId", requestId);
    body.set("websiteUrl", parsed.data.websiteUrl);
    body.set("logo", parsed.data.logo);
    try {
      const response = await fetch(
        `${apiBaseUrl || "http://localhost:8080"}/api/v1/tournaments/${tournamentId}/sponsorship-tiers/${tier.id}/checkout`,
        {
          method: "POST",
          headers: { Authorization: authorizationHeader },
          body,
        }
      );
      const result: unknown = await response.json().catch(() => null);
      if (!response.ok) {
        setSubmitError(
          extractProblemMessage(result, checkoutErrorFallback(response.status))
        );
        return;
      }
      const checkout = sponsorshipCheckoutSchema.safeParse(result);
      if (!checkout.success) {
        setSubmitError(
          "Unable to read checkout status. Retry to resume your payment."
        );
        return;
      }
      const pledge = checkout.data;
      if (["CANCELLED", "EXPIRED"].includes(pledge.status)) {
        setRequestId(crypto.randomUUID());
        setSubmitError(
          "Your previous checkout is closed. Select Continue to payment to start a new one."
        );
        return;
      }
      if (!pledge.paymentId) {
        setSubmitError(
          "Payment is still being prepared. Retry to resume the same checkout."
        );
        return;
      }
      void queryClient.invalidateQueries({
        queryKey: [
          "get",
          "/api/v1/tournaments/{tournamentId}/sponsorship-tiers",
        ],
      });
      router.push(
        `/sponsor/payments/${pledge.paymentId}/${pledge.status === "CONFIRMED" ? "receipt" : "checkout"}`
      );
      onClose();
    } catch {
      setSubmitError(
        "Connection interrupted. Retry to resume the same checkout."
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !busy) onClose();
      }}
    >
      <DialogContent
        className="sponsorship-surface max-h-[90dvh] overflow-y-auto sm:max-w-lg"
        showCloseButton={!busy}
      >
        <DialogHeader>
          <DialogTitle>Sponsor with {tier.name}</DialogTitle>
          <DialogDescription>
            {sponsorName} · {formatMoney(tier.price, currency)}. Your brand
            appears after payment succeeds.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} noValidate className="flex flex-col gap-6">
          <FieldGroup>
            <Field data-invalid={Boolean(errors.websiteUrl)}>
              <FieldLabel htmlFor="sponsor-website">Brand website</FieldLabel>
              <Input
                id="sponsor-website"
                type="url"
                value={websiteUrl}
                onChange={(event) => setWebsiteUrl(event.target.value)}
                placeholder="https://your-brand.com"
                aria-invalid={Boolean(errors.websiteUrl)}
                disabled={busy}
              />
              <FieldDescription>
                Your logo will link to this website.
              </FieldDescription>
              {errors.websiteUrl && (
                <FieldError>{errors.websiteUrl}</FieldError>
              )}
            </Field>
            <Field data-invalid={Boolean(errors.logo)}>
              <FieldLabel htmlFor="sponsor-logo">Brand logo</FieldLabel>
              <Input
                id="sponsor-logo"
                type="file"
                accept="image/png,image/jpeg"
                onChange={(event) => {
                  const file = event.target.files?.[0];
                  setLogo(file);
                  setPreview(
                    file &&
                      ["image/png", "image/jpeg"].includes(file.type) &&
                      file.size <= 1024 * 1024
                      ? URL.createObjectURL(file)
                      : ""
                  );
                }}
                aria-invalid={Boolean(errors.logo)}
                disabled={busy}
              />
              <FieldDescription>
                PNG or JPEG, up to 1 MB and 4096 × 4096 pixels.
              </FieldDescription>
              {errors.logo && <FieldError>{errors.logo}</FieldError>}
              {preview && (
                <Image
                  src={preview}
                  alt="Selected brand logo preview"
                  width={160}
                  height={96}
                  unoptimized
                  className="bg-background mt-2 h-24 w-40 rounded-md border object-contain p-2"
                />
              )}
            </Field>
          </FieldGroup>
          <p className="text-muted-foreground text-sm">
            Complete payment within 15 minutes. Verification outages may keep
            the slot reserved longer. Resume interrupted checkout from My
            payments.
          </p>
          {submitError && (
            <Alert variant="destructive">
              <AlertTitle>Checkout unavailable</AlertTitle>
              <AlertDescription>{submitError}</AlertDescription>
            </Alert>
          )}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={busy}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={busy}>
              {busy && <Spinner data-icon="inline-start" />}Continue to payment
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
