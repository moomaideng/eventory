import { NextResponse } from "next/server";
import { createClient } from "@/lib/server";
import { apiClient } from "@/lib/api/client";

export async function GET(request: Request) {
  const { searchParams, origin } = new URL(request.url);
  const code = searchParams.get("code");

  if (code) {
    try {
      const supabase = await createClient();
      const { data, error } = await supabase.auth.exchangeCodeForSession(code);

      if (!error && data?.session) {
        if (process.env.NODE_ENV === "development") {
          console.log("[Auth Callback] User Authenticated:", data.session.user.email);
        }

        // Check if account already exists in Eventory PostgreSQL DB
        try {
          const { data: account } = await apiClient.GET("/api/v1/accounts/me", {
            headers: {
              Authorization: `Bearer ${data.session.access_token}`,
            },
          });

          if (account) {
            // Case 1: Existing User -> Bypass onboarding and go straight to Home
            return NextResponse.redirect(`${origin}/`);
          }
        } catch (backendErr) {
          if (process.env.NODE_ENV === "development") {
            console.warn("[Auth Callback] Account lookup error:", backendErr);
          }
        }

        // Case 2: New User (404) -> Needs to pick display name in Onboarding
        return NextResponse.redirect(`${origin}/onboarding`);
      }
    } catch (err) {
      console.error("[Auth Callback] Auth exchange error:", err);
    }
  }

  // Fallback if exchange failed or code was missing
  return NextResponse.redirect(`${origin}/login`);
}
