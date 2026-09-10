import React from "react";
import { AuthNavbar } from "@/components/navbar";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="flex min-h-full flex-1 flex-col">
      <AuthNavbar />

      <main className="flex flex-1 flex-col">{children}</main>
    </div>
  );
}
