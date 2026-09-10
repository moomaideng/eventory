import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { createClient } from "@/lib/server";
import { LoginForm } from "@/features/auth/components/login-form";

export const metadata: Metadata = {
  title: "Sign In - Eventory",
  description: "Sign in to your Eventory account to continue.",
};

export default async function LoginPage() {
  let shouldRedirect = false;

  try {
    const supabase = await createClient();
    const { data } = await supabase.auth.getClaims();

    if (data?.claims) {
      shouldRedirect = true;
    }
  } catch {
    // Supabase client error (e.g. missing env in dev) -> allow showing login form
  }

  if (shouldRedirect) {
    redirect("/hub");
  }

  return <LoginForm />;
}
