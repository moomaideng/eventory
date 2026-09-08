import { redirect } from "next/navigation";
import { createClient } from "@/lib/server";
import { LoginForm } from "./login-form";

export default async function LoginPage() {
  let shouldRedirect = false;

  try {
    const supabase = await createClient();
    const { data } = await supabase.auth.getClaims();

    if (data?.claims) {
      shouldRedirect = true;
    }
  } catch {
    // Supabase client error (e.g. missing env in dev) -> allow showing login form
  }

  if (shouldRedirect) {
    redirect("/hub");
  }

  return <LoginForm />;
}
