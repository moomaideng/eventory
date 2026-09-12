"use client";

import { useState } from "react";
import { ShieldAlert } from "lucide-react";
import { $api, apiClient } from "@/lib/api/client";
import { useAuth } from "@/context/auth-context";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { extractProblemMessage } from "@/lib/schemas";
import type { OrganizerProfileInput } from "../schemas";
import { OrganizerProfileForm } from "./organizer-profile-form";
import { OrganizerProfileView } from "./organizer-profile-view";

interface OrganizerProfileData {
  organizerName: string;
  organizerEmail?: string | null;
}

export function OrganizerProfileManager() {
  const { isLoading: isAuthLoading, authorizationHeader } = useAuth();
  const [mode, setMode] = useState<"view" | "edit">("view");
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [override, setOverride] = useState<OrganizerProfileData | null>(
    null
  );

  const {
    data: profile,
    error,
    isLoading,
  } = $api.useQuery(
    "get",
    "/api/v1/accounts/me/organizer-profile",
    {
      headers: authorizationHeader
        ? { Authorization: authorizationHeader }
        : {},
    },
    { enabled: Boolean(authorizationHeader), retry: false, staleTime: 0 }
  );

  async function handleConfirmSave(data: OrganizerProfileInput) {
    if (!authorizationHeader) {
      setSaveError("Sign in before saving your organizer profile.");
      return false;
    }

    setSaveError("");
    setIsSaving(true);
    try {
      const { data: putData, error: putError } = await apiClient.PUT(
        "/api/v1/accounts/me/organizer-profile",
        {
          headers: { Authorization: authorizationHeader },
          body: {
            organizerName: data.organizerName,
            organizerEmail: data.organizerEmail,
          },
        }
      );

      if (putError || !putData) {
        setSaveError(
          extractProblemMessage(
            putError,
            "We could not save your organizer profile."
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

  if (error || !current) {
    return (
      <Empty className="border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ShieldAlert />
          </EmptyMedia>
          <EmptyTitle>Organizer profile unavailable</EmptyTitle>
          <EmptyDescription>
            We could not load your organizer profile. Try refreshing the
            page.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  if (mode === "edit") {
    return (
      <OrganizerProfileForm
        initialName={current.organizerName}
        initialEmail={current.organizerEmail ?? ""}
        onCancel={() => setMode("view")}
        onConfirmSave={handleConfirmSave}
        isSaving={isSaving}
        saveError={saveError}
      />
    );
  }

  return (
    <OrganizerProfileView
      organizerName={current.organizerName}
      organizerEmail={current.organizerEmail}
      onEdit={() => setMode("edit")}
    />
  );
}
