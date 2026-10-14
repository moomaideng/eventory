"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertCircle } from "lucide-react";
import { useAuth } from "@/context/auth-context";
import { $api } from "@/lib/api/client";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
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
import { Textarea } from "@/components/ui/textarea";
import {
  sponsorshipTierSchema,
  type SponsorshipTierFormInput,
} from "@/features/sponsorship/schemas";
import {
  getErrorMessage,
  type SponsorshipTier,
} from "@/features/sponsorship/utils";

const EMPTY_FORM: SponsorshipTierFormInput = {
  name: "",
  price: "",
  slots: "",
  perks: "",
  sortOrder: 0,
};

function toFormValues(tier: SponsorshipTier | null): SponsorshipTierFormInput {
  if (!tier) return EMPTY_FORM;
  return {
    name: tier.name,
    price: String(tier.price),
    slots: String(tier.slots),
    perks: (tier.perks ?? []).join("\n"),
    sortOrder: tier.sortOrder,
  };
}

/** Creates a tier when `tier` is null, otherwise edits it. */
export function TierFormDialog({
  tournamentId,
  currency,
  tier,
  open,
  onOpenChange,
}: {
  tournamentId: string;
  currency: string;
  tier: SponsorshipTier | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [values, setValues] = useState(() => toFormValues(tier));
  const [submitted, setSubmitted] = useState(false);
  const { authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  const createMutation = $api.useMutation(
    "post",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers"
  );
  const updateMutation = $api.useMutation(
    "put",
    "/api/v1/tournaments/{tournamentId}/sponsorship-tiers/{tierId}"
  );
  const mutation = tier ? updateMutation : createMutation;

  const parsed = sponsorshipTierSchema.safeParse(values);
  const fieldErrors =
    submitted && !parsed.success ? parsed.error.flatten().fieldErrors : {};

  function setField(key: keyof SponsorshipTierFormInput, value: string) {
    setValues((current) => ({ ...current, [key]: value }));
  }

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitted(true);
    if (!parsed.success) return;

    const options = {
      onSuccess: () => {
        void queryClient.invalidateQueries();
        onOpenChange(false);
      },
    };
    const headers = { Authorization: authorizationHeader ?? "" };
    if (tier) {
      updateMutation.mutate(
        {
          params: { path: { tournamentId, tierId: tier.id } },
          headers,
          body: parsed.data,
        },
        options
      );
    } else {
      createMutation.mutate(
        { params: { path: { tournamentId } }, headers, body: parsed.data },
        options
      );
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6">
          <DialogHeader>
            <DialogTitle>
              {tier ? "Edit sponsorship tier" : "New sponsorship tier"}
            </DialogTitle>
            <DialogDescription>
              Sponsors buy tiers from your public tournament page.
            </DialogDescription>
          </DialogHeader>

          <FieldGroup>
            <Field data-invalid={fieldErrors.name ? true : undefined}>
              <FieldLabel htmlFor="tier-name">Name</FieldLabel>
              <Input
                id="tier-name"
                placeholder="Gold"
                value={values.name}
                onChange={(e) => setField("name", e.target.value)}
                aria-invalid={fieldErrors.name ? true : undefined}
                autoComplete="off"
              />
              {fieldErrors.name ? (
                <FieldError>{fieldErrors.name[0]}</FieldError>
              ) : null}
            </Field>

            <div className="grid grid-cols-2 gap-4">
              <Field data-invalid={fieldErrors.price ? true : undefined}>
                <FieldLabel htmlFor="tier-price">Price ({currency})</FieldLabel>
                <Input
                  id="tier-price"
                  type="number"
                  inputMode="numeric"
                  min={1}
                  value={String(values.price)}
                  onChange={(e) => setField("price", e.target.value)}
                  aria-invalid={fieldErrors.price ? true : undefined}
                />
                {fieldErrors.price ? (
                  <FieldError>{fieldErrors.price[0]}</FieldError>
                ) : null}
              </Field>

              <Field data-invalid={fieldErrors.slots ? true : undefined}>
                <FieldLabel htmlFor="tier-slots">Slots</FieldLabel>
                <Input
                  id="tier-slots"
                  type="number"
                  inputMode="numeric"
                  min={1}
                  value={String(values.slots)}
                  onChange={(e) => setField("slots", e.target.value)}
                  aria-invalid={fieldErrors.slots ? true : undefined}
                />
                {fieldErrors.slots ? (
                  <FieldError>{fieldErrors.slots[0]}</FieldError>
                ) : null}
              </Field>
            </div>

            <Field data-invalid={fieldErrors.perks ? true : undefined}>
              <FieldLabel htmlFor="tier-perks">Perks</FieldLabel>
              <Textarea
                id="tier-perks"
                rows={4}
                placeholder={"Logo on stream\nBooth at the venue"}
                value={String(values.perks)}
                onChange={(e) => setField("perks", e.target.value)}
                aria-invalid={fieldErrors.perks ? true : undefined}
              />
              {fieldErrors.perks ? (
                <FieldError>{fieldErrors.perks[0]}</FieldError>
              ) : (
                <FieldDescription>One perk per line, up to 10.</FieldDescription>
              )}
            </Field>

            <Field>
              <FieldLabel htmlFor="tier-order">Display order</FieldLabel>
              <Input
                id="tier-order"
                type="number"
                inputMode="numeric"
                value={String(values.sortOrder)}
                onChange={(e) => setField("sortOrder", e.target.value)}
              />
              <FieldDescription>Lower numbers show first.</FieldDescription>
            </Field>
          </FieldGroup>

          {mutation.isError ? (
            <Alert variant="destructive">
              <AlertCircle />
              <AlertTitle>Could not save tier</AlertTitle>
              <AlertDescription>{getErrorMessage(mutation.error)}</AlertDescription>
            </Alert>
          ) : null}

          <DialogFooter>
            <DialogClose
              render={<Button variant="outline" type="button" />}
              disabled={mutation.isPending}
            >
              Cancel
            </DialogClose>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? <Spinner data-icon="inline-start" /> : null}
              {tier ? "Save changes" : "Create tier"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
