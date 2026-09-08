import { redirect } from "next/navigation";
import { createClient } from "@/lib/server";
import { LoginForm } from "./login-form";

export default async function LoginPage() {
  try {
    const supabase = await createClient();
    const {
      data: { user },
    } = await supabase.auth.getUser();

    if (user) {
      redirect("/");
    }
  } catch {
    // Supabase client error (e.g. missing env in dev) -> allow showing login form
  }

  return <LoginForm />;
}
