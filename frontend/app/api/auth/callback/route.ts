import { NextResponse } from "next/server";
import { createClient } from "@/lib/server";

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

        // Successfully authenticated with Supabase, proceed to onboarding
        return NextResponse.redirect(`${origin}/onboarding`);
      }
    } catch (err) {
      console.error("[Auth Callback] Auth exchange error:", err);
    }
  }

  // Fallback if exchange failed or code was missing
  return NextResponse.redirect(`${origin}/login`);
}
