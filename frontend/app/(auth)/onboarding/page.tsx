import { redirect } from "next/navigation";
import { createClient } from "@/lib/server";
import { apiClient } from "@/lib/api/client";
import { OnboardingForm } from "./onboarding-form";

export default async function OnboardingPage() {
  let shouldRedirectToHome = false;
  let shouldRedirectToLogin = false;

  try {
    const supabase = await createClient();
    const {
      data: { session },
    } = await supabase.auth.getSession();

    if (!session?.access_token) {
      shouldRedirectToLogin = true;
    } else {
      try {
        const { data: account } = await apiClient.GET("/api/v1/accounts/me", {
          headers: {
            Authorization: `Bearer ${session.access_token}`,
          },
        });

        if (account) {
          shouldRedirectToHome = true;
        }
      } catch {
        // Backend not reachable or error -> allow showing onboarding form
      }
    }
  } catch {
    // Supabase client creation error (e.g. missing env in test/local)
  }

  if (shouldRedirectToHome) {
    redirect("/");
  }

  if (shouldRedirectToLogin) {
    redirect("/login");
  }

  return <OnboardingForm />;
}
