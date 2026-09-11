"use client";

import React from "react";
import { Building2 } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { sponsorProfileSchema } from "../schemas";

type FieldErrors = Partial<Record<"sponsorName" | "sponsorEmail", string>>;

export function SponsorProfileForm() {
  const [sponsorName, setSponsorName] = React.useState("");
  const [sponsorEmail, setSponsorEmail] = React.useState("");
  const [fieldErrors, setFieldErrors] = React.useState<FieldErrors>({});
  const [saved, setSaved] = React.useState(false);

  function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaved(false);

    const result = sponsorProfileSchema.safeParse({
      sponsorName,
      sponsorEmail,
    });

    if (!result.success) {
      const flat = result.error.flatten().fieldErrors;
      setFieldErrors({
        sponsorName: flat.sponsorName?.[0],
        sponsorEmail: flat.sponsorEmail?.[0],
      });
      return;
    }

    setFieldErrors({});
    // TODO(#64): call apiClient.PUT("/api/v1/accounts/me/sponsor-profile", ...)
    // with result.data once API wiring lands. For now this only confirms the
    // form validates correctly.
    setSaved(true);
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Sponsor profile</CardTitle>
        <CardDescription>
          This information is shown to organizers when you sponsor a
          tournament.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit} noValidate>
        <CardContent>
          <FieldGroup>
            <Field data-invalid={Boolean(fieldErrors.sponsorName)}>
              <FieldLabel htmlFor="sponsor-name">Sponsor name</FieldLabel>
              <Input
                id="sponsor-name"
                value={sponsorName}
                onChange={(event) => {
                  setSponsorName(event.target.value);
                  setFieldErrors((prev) => ({
                    ...prev,
                    sponsorName: undefined,
                  }));
                }}
                placeholder="Acme Corp"
                maxLength={120}
                aria-invalid={Boolean(fieldErrors.sponsorName)}
              />
              {fieldErrors.sponsorName ? (
                <FieldError>{fieldErrors.sponsorName}</FieldError>
              ) : null}
            </Field>

            <Field data-invalid={Boolean(fieldErrors.sponsorEmail)}>
              <FieldLabel htmlFor="sponsor-email">Contact email</FieldLabel>
              <Input
                id="sponsor-email"
                type="email"
                value={sponsorEmail}
                onChange={(event) => {
                  setSponsorEmail(event.target.value);
                  setFieldErrors((prev) => ({
                    ...prev,
                    sponsorEmail: undefined,
                  }));
                }}
                placeholder="sponsorships@acme.com"
                aria-invalid={Boolean(fieldErrors.sponsorEmail)}
              />
              <FieldDescription>Optional.</FieldDescription>
              {fieldErrors.sponsorEmail ? (
                <FieldError>{fieldErrors.sponsorEmail}</FieldError>
              ) : null}
            </Field>
          </FieldGroup>
        </CardContent>
        <CardFooter className="justify-end gap-3">
          {saved ? (
            <span className="text-muted-foreground text-sm">
              Looks good — saving isn&apos;t wired up yet.
            </span>
          ) : null}
          <Button type="submit">
            <Building2 data-icon="inline-start" />
            Save profile
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
}
