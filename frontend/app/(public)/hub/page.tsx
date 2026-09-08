import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { createClient } from "@/lib/server";
import { ModeHub } from "./mode-hub";

export const metadata: Metadata = {
  title: "Choose Your Mode - Eventory",
  description:
    "Select what profile mode you want to use to interact with Eventory.",
};

export default async function HubPage() {
  let shouldRedirect = false;

  try {
    const supabase = await createClient();
    const { data } = await supabase.auth.getClaims();

    if (!data?.claims) {
      shouldRedirect = true;
    }
  } catch {
    // In dev offline mode or if client fails, allow ModeHub to render with mock/dev states
  }

  if (shouldRedirect) {
    redirect("/login");
  }

  return <ModeHub />;
}
