"use client";

import { useState } from "react";
import { Plus, ShieldAlert } from "lucide-react";
import { $api, apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { extractProblemMessage } from "@/lib/schemas";
import type { SponsorProfileInput } from "../schemas";
import { SponsorProfileForm } from "./sponsor-profile-form";
import { SponsorProfileView } from "./sponsor-profile-view";

interface SponsorProfileData {
  sponsorName: string;
  sponsorEmail?: string | null;
  phone?: string | null;
  tags?: string[] | null;
}

export function SponsorProfileManager() {
  const { isLoading: isAuthLoading, authorizationHeader } = useAuth();
  const [mode, setMode] = useState<"view" | "edit">("view");
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [override, setOverride] = useState<SponsorProfileData | null>(null);

  const {
    data: profile,
    error,
    isLoading,
  } = $api.useQuery(
    "get",
    "/api/v1/accounts/me/sponsor-profile",
    {
      headers: authorizationHeader
        ? { Authorization: authorizationHeader }
        : {},
    },
    { enabled: Boolean(authorizationHeader), retry: false, staleTime: 0 }
  );

  async function handleConfirmSave(data: SponsorProfileInput) {
    if (!authorizationHeader) {
      setSaveError("Sign in before saving your sponsor profile.");
      return false;
    }

    setSaveError("");
    setIsSaving(true);
    try {
      const { data: putData, error: putError } = await apiClient.PUT(
        "/api/v1/accounts/me/sponsor-profile",
        {
          headers: { Authorization: authorizationHeader },
          body: {
            sponsorName: data.sponsorName,
            sponsorEmail: data.sponsorEmail,
            phone: data.phone,
            tags: data.tags,
          },
        }
      );

      if (putError || !putData) {
        setSaveError(
          extractProblemMessage(
            putError,
            "We could not save your sponsor profile."
          )
        );
        return false;
      }

      setOverride(putData);
      setMode("view");
      return true;
    } catch {
      setSaveError(
        "Connection error. Please check your network and try again."
      );
      return false;
    } finally {
      setIsSaving(false);
    }
  }

  if (isAuthLoading || isLoading) {
    return (
      <div className="flex flex-col gap-6">
        <Skeleton className="h-12 w-72" />
        <Skeleton className="h-64 w-full rounded-xl" />
      </div>
    );
  }

  const current = override ?? profile;
  const hasNoProfile = !current && error?.status === 404;

  if (!current && !hasNoProfile) {
    return (
      <Empty className="border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ShieldAlert />
          </EmptyMedia>
          <EmptyTitle>Sponsor profile unavailable</EmptyTitle>
          <EmptyDescription>
            We could not load your sponsor profile. Try refreshing the page.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  if (mode === "edit") {
    return (
      <SponsorProfileForm
        initialName={current?.sponsorName ?? ""}
        initialEmail={current?.sponsorEmail ?? ""}
        initialPhone={current?.phone ?? ""}
        initialTags={current?.tags ?? []}
        onCancel={() => setMode("view")}
        onConfirmSave={handleConfirmSave}
        isSaving={isSaving}
        saveError={saveError}
      />
    );
  }

  if (!current) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Sponsor profile</CardTitle>
          <CardDescription>
            You do not have a sponsor profile yet. Create one to appear in the
            sponsor directory.
          </CardDescription>
        </CardHeader>
        <CardFooter className="justify-end">
          <Button onClick={() => setMode("edit")}>
            <Plus data-icon="inline-start" />
            Create sponsor profile
          </Button>
        </CardFooter>
      </Card>
    );
  }

  return (
    <SponsorProfileView
      sponsorName={current.sponsorName}
      sponsorEmail={current.sponsorEmail}
      phone={current.phone}
      tags={current.tags ?? []}
      onEdit={() => setMode("edit")}
    />
  );
}
