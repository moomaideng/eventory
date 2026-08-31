import { redirect } from "next/navigation";
import { createClient } from "@/lib/server";
import { apiClient } from "@/lib/api/client";
import { LoginForm } from "./login-form";

export default async function LoginPage() {
  let shouldRedirectToHome = false;
  let shouldRedirectToOnboarding = false;

  try {
    const supabase = await createClient();
    const {
      data: { session },
    } = await supabase.auth.getSession();

    if (session?.access_token) {
      try {
        const { data: account } = await apiClient.GET("/api/v1/accounts/me", {
          headers: {
            Authorization: `Bearer ${session.access_token}`,
          },
        });

        if (account) {
          shouldRedirectToHome = true;
        } else {
          shouldRedirectToOnboarding = true;
        }
      } catch {
        // Backend not reachable or error -> allow showing login form
      }
    }
  } catch {
    // Supabase client error (e.g. missing env in dev) -> allow showing login form
  }

  if (shouldRedirectToHome) {
    redirect("/");
  }

  if (shouldRedirectToOnboarding) {
    redirect("/onboarding");
  }

  return <LoginForm />;
}
