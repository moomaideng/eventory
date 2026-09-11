import { Suspense } from "react";
import type { Metadata } from "next";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { createClient } from "@/lib/server";
import { LoginForm } from "@/features/auth/components/login-form";
import { getSafeRedirectPath } from "@/lib/auth-redirect";
import AuthLoading from "@/app/(auth)/loading";

export const metadata: Metadata = {
  title: "Sign In - Eventory",
  description: "Sign in to your Eventory account to continue.",
};

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ redirectTo?: string }>;
}) {
  const { redirectTo } = await searchParams;
  const safeRedirectTo = getSafeRedirectPath(redirectTo);
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

  // [DEV-ONLY] Redirect if dev session is already active
  if (process.env.NODE_ENV === "development") {
    const cookieStore = await cookies();
    if (cookieStore.get("eventory_dev_session")?.value === "true") {
      shouldRedirect = true;
    }
  }

  if (shouldRedirect) {
    redirect(safeRedirectTo);
  }

  return (
    <Suspense fallback={<AuthLoading />}>
      <LoginForm />
    </Suspense>
  );
}
