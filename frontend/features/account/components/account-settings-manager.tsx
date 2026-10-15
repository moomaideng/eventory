"use client";

import { ShieldAlert } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import {
  AccountSettingsForm,
  type SaveResult,
} from "@/features/account/components/account-settings-form";
import type { AccountSettingsInput } from "@/features/account/schemas";
import { useAuth } from "@/context/auth-context";
import { $api, apiClient } from "@/lib/api/client";
import type { components } from "@/lib/api/schema";
import { extractProblemMessage } from "@/lib/schemas";

type Account = components["schemas"]["AccountResponse"];

// Matches every cached copy of GET /accounts/me, whatever its request params.
const ACCOUNT_QUERY_KEY = ["get", "/api/v1/accounts/me"] as const;

// The API leaves empty phone/avatar out (or sends null); the form uses "".
function toFormValues(account: Account): AccountSettingsInput {
  return {
    displayName: account.displayName,
    handle: account.handle,
    phone: account.phone ?? "",
    avatarUrl: account.avatarUrl ?? "",
  };
}

export function AccountSettingsManager() {
  const { isLoading: isAuthLoading, authorizationHeader } = useAuth();
  const queryClient = useQueryClient();

  // Same request as auth-context, so this reads the shared cache entry
  // instead of calling the backend a second time.
  const {
    data: account,
    error,
    isLoading,
  } = $api.useQuery(
    "get",
    "/api/v1/accounts/me",
    { headers: { Authorization: authorizationHeader ?? "" } },
    {
      enabled: Boolean(authorizationHeader),
      retry: false,
      staleTime: 60 * 1000,
    }
  );

  async function handleSave(values: AccountSettingsInput): Promise<SaveResult> {
    if (!authorizationHeader) {
      return { ok: false, message: "Sign in again to save your changes." };
    }

    try {
      const {
        data,
        error: patchError,
        response,
      } = await apiClient.PATCH("/api/v1/accounts/me", {
        headers: { Authorization: authorizationHeader },
        body: values,
      });

      if (patchError || !data) {
        const message = extractProblemMessage(
          patchError,
          "We could not save your changes. Please try again."
        );
        // 409 means the handle belongs to someone else.
        return response.status === 409
          ? { ok: false, field: "handle", message }
          : { ok: false, message };
      }

      // Put the saved account in every cached copy (the navbar reads it too),
      // then refetch in the background to confirm.
      queryClient.setQueriesData<Account>(
        { queryKey: ACCOUNT_QUERY_KEY },
        data
      );
      queryClient.invalidateQueries({ queryKey: ACCOUNT_QUERY_KEY });

      return { ok: true, values: toFormValues(data) };
    } catch {
      return {
        ok: false,
        message: "Connection error. Please check your network and try again.",
      };
    }
  }

  if (isAuthLoading || isLoading) {
    return (
      <div className="flex flex-col gap-6">
        <Skeleton className="h-24 w-full rounded-xl" />
        <Skeleton className="h-96 w-full rounded-xl" />
      </div>
    );
  }

  if (error || !account) {
    return (
      <Empty className="border">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ShieldAlert />
          </EmptyMedia>
          <EmptyTitle>Account unavailable</EmptyTitle>
          <EmptyDescription>
            We could not load your account. Try refreshing the page.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    );
  }

  return (
    <AccountSettingsForm
      key={account.id}
      initial={toFormValues(account)}
      onSave={handleSave}
    />
  );
}
