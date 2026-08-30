import { NextResponse } from "next/server";
import { createClient } from "@/lib/server";

export async function GET(request: Request) {
  const { searchParams, origin } = new URL(request.url);
  const code = searchParams.get("code");

  if (code) {
    try {
      const supabase = await createClient();
      const { error } = await supabase.auth.exchangeCodeForSession(code);

      if (!error) {
        // Successfully exchanged code for session, redirect to Onboarding
        return NextResponse.redirect(`${origin}/onboarding`);
      }
    } catch (err) {
      console.error("Auth exchange error:", err);
    }
  }

  // Fallback if exchange failed or code was missing
  return NextResponse.redirect(`${origin}/login`);
}
