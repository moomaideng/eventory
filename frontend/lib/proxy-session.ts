import { createServerClient } from "@supabase/ssr";
import { NextResponse, type NextRequest } from "next/server";

export interface ProxySessionResult {
  response: NextResponse;
  user: { id: string } | null;
}

export async function updateSession(
  request: NextRequest
): Promise<ProxySessionResult> {
  let supabaseResponse = NextResponse.next({
    request,
  });

  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL;
  const supabaseKey = process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY;

  // Safe fallback for dev offline mode when Supabase credentials are not configured
  if (!supabaseUrl || !supabaseKey) {
    return {
      response: supabaseResponse,
      user: { id: "offline-dev" },
    };
  }

  // Support dev session bypass in development mode
  if (
    process.env.NODE_ENV === "development" &&
    request.cookies.get("eventory_dev_session")?.value === "true"
  ) {
    return {
      response: supabaseResponse,
      user: { id: "dev-user" },
    };
  }

  let user: { id: string } | null = null;

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
    const { data } = await supabase.auth.getClaims();
    if (data?.claims?.sub) {
      user = { id: data.claims.sub };
    }
  } catch (err) {
    if (process.env.NODE_ENV === "development") {
      console.warn("[Proxy] Session refresh notice:", err);
      // Safe fallback for dev offline mode when local Supabase instance is unreachable
      return {
        response: supabaseResponse,
        user: { id: "dev-offline-fallback" },
      };
    }
  }

  return {
    response: supabaseResponse,
    user,
  };
}
