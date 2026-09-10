import { NextResponse, type NextRequest } from "next/server";
import { updateSession } from "@/lib/proxy-session";

const PROTECTED_PREFIXES = [
  "/hub",
  "/organizer",
  "/sponsor",
  "/settings",
  "/lobbies",
];

function isPathProtected(pathname: string): boolean {
  if (
    PROTECTED_PREFIXES.some(
      (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`)
    )
  ) {
    return true;
  }
  // Protect tournament team creation route: /tournaments/:tournamentId/team
  if (/^\/tournaments\/[^/]+\/team(\/.*)?$/.test(pathname)) {
    return true;
  }
  return false;
}

/**
 * Next.js 16 Proxy Convention
 * Handles centralized route protection and Supabase SSR session token refresh.
 */
export async function proxy(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const { response, user } = await updateSession(request);

  const isProtected = isPathProtected(pathname);

  if (isProtected && !user) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("redirectTo", pathname + search);
    const redirectResponse = NextResponse.redirect(loginUrl);

    // Forward any session cookie updates/deletions from Supabase SSR
    response.cookies.getAll().forEach((cookie) => {
      redirectResponse.cookies.set(cookie);
    });

    return redirectResponse;
  }

  return response;
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - static image formats (.svg, .png, .jpg, .jpeg, .gif, .webp)
     */
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)",
  ],
};
