import React from "react";
import { redirect } from "next/navigation";
import { Navbar } from "@/components/navbar";
import { createClient } from "@/lib/server";
import { apiClient } from "@/lib/api/client";

export default async function PublicLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
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

        // Strict Mode: If user is authenticated with Supabase but has not created an account in DB yet, enforce onboarding
        if (!account) {
          shouldRedirectToOnboarding = true;
        }
      } catch {
        // If backend is unreachable or account is 404, enforce onboarding
        shouldRedirectToOnboarding = true;
      }
    }
  } catch {
    // Supabase client error in offline/dev
  }

  if (shouldRedirectToOnboarding) {
    redirect("/onboarding");
  }

  return (
    <>
      <Navbar />
      <main className="flex flex-1 flex-col">{children}</main>
    </>
  );
}
