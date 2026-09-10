"use server";

import { cookies } from "next/headers";

/**
 * Dev Session Cookie Management (Server Actions)
 *
 * Secure HttpOnly cookie handler for local offline development mode.
 * - Follows React Doctor security best practices by managing auth cookies server-side with `httpOnly: true`.
 * - Production authentication uses Supabase SSR; this helper is strictly for local dev convenience.
 */

const DEV_SESSION_COOKIE = "eventory_dev_session";

function checkAuth() {
  if (process.env.NODE_ENV === "production") {
    throw new Error("Unauthorized: Dev session is only available in development environment.");
  }
}

export async function setDevSessionAction(enabled: boolean): Promise<void> {
  checkAuth();
  const cookieStore = await cookies();
  if (enabled) {
    cookieStore.set(DEV_SESSION_COOKIE, "true", {
      httpOnly: true,
      secure: process.env.NODE_ENV === "production",
      sameSite: "lax",
      path: "/",
      maxAge: 86400, // 24 hours
    });
  } else {
    cookieStore.delete(DEV_SESSION_COOKIE);
    cookieStore.delete("eventory_dev_role");
  }
}

export async function getDevSessionAction(): Promise<boolean> {
  const cookieStore = await cookies();
  return cookieStore.get(DEV_SESSION_COOKIE)?.value === "true";
}
