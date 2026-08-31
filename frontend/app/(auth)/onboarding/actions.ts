"use server";

import { createClient } from "@/lib/server";
import { apiClient } from "@/lib/api/client";

export interface OnboardResult {
  success?: boolean;
  error?: string;
}

export async function onboardUser(username: string): Promise<OnboardResult> {
  const trimmed = username.trim();
  if (!trimmed) {
    return { error: "Please enter a valid display name." };
  }

  try {
    const supabase = await createClient();
    const {
      data: { session },
    } = await supabase.auth.getSession();

    if (!session?.access_token) {
      return { error: "Authentication session expired. Please log in again." };
    }

    // Call Go Backend Onboarding API with server-held access token
    const { error } = await apiClient.POST("/api/v1/accounts/onboard", {
      body: {
        username: trimmed,
      },
      headers: {
        Authorization: `Bearer ${session.access_token}`,
      },
    });

    if (error) {
      return {
        error:
          error.detail ||
          "Failed to create account. Please try another username.",
      };
    }

    return { success: true };
  } catch (err) {
    console.error("Onboarding server action error:", err);
    return { error: "Unable to connect to server. Please try again." };
  }
}
