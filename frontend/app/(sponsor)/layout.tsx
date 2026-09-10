import React from "react";
import { SponsorNavbar } from "@/components/navbar/sponsor-navbar";

export default function SponsorLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <SponsorNavbar />
      <main className="flex flex-1 flex-col">{children}</main>
    </>
  );
}
