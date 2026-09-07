import { NextResponse } from "next/server";
import { createClient } from "@/lib/server";
import { apiClient } from "@/lib/api/client";

function getPublicOrigin(request: Request) {
  const forwardedHost = request.headers.get("x-forwarded-host");
  const host = forwardedHost || request.headers.get("host");
  const forwardedProtocol = request.headers
    .get("x-forwarded-proto")
    ?.split(",")[0];

  if (host) {
    const protocol =
      forwardedProtocol || new URL(request.url).protocol.slice(0, -1);
    return `${protocol}://${host}`;
  }

  return new URL(request.url).origin;
}

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const origin = getPublicOrigin(request);
  const code = searchParams.get("code");

  if (code) {
    try {
      const supabase = await createClient();
      const { data, error } = await supabase.auth.exchangeCodeForSession(code);

      if (!error && data?.session) {
        if (process.env.NODE_ENV === "development") {
          console.log(
            "[Auth Callback] User Authenticated:",
            data.session.user.email
          );
        }

        // Just-In-Time (JIT) Auto-Provision Account in Eventory PostgreSQL DB
        try {
          const userMeta = data.session.user.user_metadata || {};
          const displayName =
            userMeta.full_name ||
            userMeta.name ||
            data.session.user.email?.split("@")[0] ||
            "User";
          const avatarUrl = userMeta.avatar_url || userMeta.picture;

          await apiClient.POST("/api/v1/accounts", {
            headers: {
              Authorization: `Bearer ${data.session.access_token}`,
            },
            body: {
              displayName,
              avatarUrl,
            },
          });
        } catch (backendErr) {
          if (process.env.NODE_ENV === "development") {
            console.warn(
              "[Auth Callback] Account JIT auto-provisioning error:",
              backendErr
            );
          }
        }

        return NextResponse.redirect(`${origin}/`);
      }
    } catch (err) {
      console.error("[Auth Callback] Auth exchange error:", err);
    }
  }

  // Fallback if exchange failed or code was missing
  return NextResponse.redirect(`${origin}/login`);
}
