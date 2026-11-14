"use server";

import { cookies } from "next/headers";
import {
  parseDevSessionCookie,
  type DevProfileId,
} from "@/context/mock-data";

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
    throw new Error(
      "Unauthorized: Dev session is only available in development environment."
    );
  }
}

export async function setDevSessionAction(
  profile: DevProfileId | boolean
): Promise<void> {
  checkAuth();
  const cookieStore = await cookies();
  if (profile) {
    const cookieValue: DevProfileId =
      profile === true ? "alex-dev" : profile;
    cookieStore.set(DEV_SESSION_COOKIE, cookieValue, {
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

export async function getDevSessionAction(): Promise<DevProfileId | null> {
  const cookieStore = await cookies();
  return parseDevSessionCookie(cookieStore.get(DEV_SESSION_COOKIE)?.value);
}
