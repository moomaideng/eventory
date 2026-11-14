"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { useAuth } from "@/context/auth-context";
import { apiClient } from "@/lib/api/client";
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
import { Spinner } from "@/components/ui/spinner";
import { sponsorshipCheckoutSchema } from "@/features/sponsorship/schemas";
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
  websiteUrl,
  logoUrl,
  onClose,
}: {
  tournamentId: string;
  tier: SponsorshipTier;
  currency: string;
  sponsorName: string;
  websiteUrl: string;
  logoUrl: string;
  onClose: () => void;
}) {
  const { authorizationHeader } = useAuth();
  const router = useRouter();
  const queryClient = useQueryClient();
  const [submitError, setSubmitError] = useState("");
  const [busy, setBusy] = useState(false);
  const [requestId, setRequestId] = useState(() => crypto.randomUUID());

  const hasBrandAssets = Boolean(websiteUrl.trim() && logoUrl.trim());

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy || !authorizationHeader || !hasBrandAssets) return;
    setSubmitError("");
    setBusy(true);
    try {
      const { data, error, response } = await apiClient.POST(
        "/api/v1/tournaments/{tournamentId}/sponsorship-tiers/{tierId}/checkout",
        {
          params: { path: { tournamentId, tierId: tier.id } },
          headers: { Authorization: authorizationHeader },
          body: { requestId },
        }
      );
      if (error || !data) {
        setSubmitError(
          extractProblemMessage(error, checkoutErrorFallback(response.status))
        );
        return;
      }
      const checkout = sponsorshipCheckoutSchema.safeParse(data);
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
        `/payments/${pledge.paymentId}/${pledge.status === "CONFIRMED" ? "receipt" : "checkout"}`
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
            {sponsorName} · {formatMoney(tier.priceMinor, currency)}. Your
            profile brand and logo appear on the tournament page after payment
            succeeds.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} noValidate className="flex flex-col gap-5">
          {hasBrandAssets ? (
            <div className="border-border bg-muted/30 flex items-center gap-4 rounded-lg border p-3">
              <Image
                src={logoUrl}
                alt={`${sponsorName} logo`}
                width={96}
                height={64}
                unoptimized
                className="bg-background h-16 w-24 shrink-0 rounded-md object-contain p-1"
              />
              <div className="min-w-0 flex-1 text-sm">
                <p className="font-semibold wrap-anywhere">{sponsorName}</p>
                <p className="text-muted-foreground truncate">{websiteUrl}</p>
              </div>
            </div>
          ) : (
            <Alert variant="destructive">
              <AlertTitle>Brand profile incomplete</AlertTitle>
              <AlertDescription className="flex flex-col gap-3">
                <span>
                  Add your brand website and logo image URL in Sponsor Profile
                  settings before checking out.
                </span>
                <div>
                  <Button
                    size="sm"
                    variant="outline"
                    render={<Link href="/settings" />}
                    nativeButton={false}
                  >
                    Open Sponsor Profile
                  </Button>
                </div>
              </AlertDescription>
            </Alert>
          )}
          <p className="text-muted-foreground text-sm">
            Complete payment within 15 minutes. Resume interrupted checkout from
            My payments.
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
            <Button type="submit" disabled={busy || !hasBrandAssets}>
              {busy && <Spinner data-icon="inline-start" />}Continue to payment
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
