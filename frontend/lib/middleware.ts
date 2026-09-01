import { createServerClient } from "@supabase/ssr";
import { NextResponse, type NextRequest } from "next/server";

export async function updateSession(request: NextRequest) {
  let supabaseResponse = NextResponse.next({
    request,
  });

  const supabaseUrl = process.env.SUPABASE_URL;
  const supabaseKey = process.env.SUPABASE_PUBLISHABLE_KEY;

  // If Supabase environment variables are not set (e.g. local offline dev), pass through gracefully
  if (!supabaseUrl || !supabaseKey) {
    return supabaseResponse;
  }

  try {
    const supabase = createServerClient(supabaseUrl, supabaseKey, {
      cookies: {
        getAll() {
          return request.cookies.getAll();
        },
        setAll(cookiesToSet) {
          cookiesToSet.forEach(({ name, value }) =>
            request.cookies.set(name, value)
          );
          supabaseResponse = NextResponse.next({
            request,
          });
          cookiesToSet.forEach(({ name, value, options }) =>
            supabaseResponse.cookies.set(name, value, options)
          );
        },
      },
    });

    // With modern Asymmetric JWT Signing, getClaims() performs local cryptographic verification (via JWKS)
    // without a network roundtrip to Supabase Auth, while keeping session cookies automatically refreshed.
    await supabase.auth.getClaims();
  } catch (err) {
    if (process.env.NODE_ENV === "development") {
      console.warn("[Proxy / Middleware] Session refresh notice:", err);
    }
  }

  return supabaseResponse;
}
