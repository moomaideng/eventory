import { NextResponse } from "next/server";
import { createClient } from "@/lib/server";
import { apiClient } from "@/lib/api/client";
import { getSafeRedirectPath } from "@/lib/auth-redirect";
import {
  authCallbackQuerySchema,
  authUserMetadataSchema,
  accountProvisionSchema,
} from "@/features/auth/schemas";

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

  const queryResult = authCallbackQuerySchema.safeParse({
    code: searchParams.get("code") ?? undefined,
    error: searchParams.get("error") ?? undefined,
    error_description: searchParams.get("error_description") ?? undefined,
  });

  if (queryResult.success && queryResult.data.code) {
    const { code } = queryResult.data;
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
          const rawMeta = data.session.user.user_metadata || {};
          const metaResult = authUserMetadataSchema.safeParse(rawMeta);
          const userMeta = metaResult.success ? metaResult.data : {};

          const fallbackName =
            data.session.user.email?.split("@")[0] || "User";
          const rawDisplayName =
            userMeta.full_name || userMeta.name || fallbackName;
          const rawAvatarUrl = userMeta.avatar_url || userMeta.picture;

          const provisionResult = accountProvisionSchema.safeParse({
            displayName: rawDisplayName,
            avatarUrl: rawAvatarUrl || undefined,
          });

          if (provisionResult.success) {
            await apiClient.POST("/api/v1/accounts", {
              headers: {
                Authorization: `Bearer ${data.session.access_token}`,
              },
              body: provisionResult.data,
            });
          }
        } catch (backendErr) {
          if (process.env.NODE_ENV === "development") {
            console.warn(
              "[Auth Callback] Account JIT auto-provisioning error:",
              backendErr
            );
          }
        }

        const next = searchParams.get("next");
        const safeNext = getSafeRedirectPath(next);
        return NextResponse.redirect(`${origin}${safeNext}`);
      }
    } catch (err) {
      console.error("[Auth Callback] Auth exchange error:", err);
    }
  }

  // Fallback if exchange failed or code was missing
  return NextResponse.redirect(`${origin}/login`);
}
