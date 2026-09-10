import React from "react";
import { OrganizerNavbar } from "@/components/navbar/organizer-navbar";

export default function OrganizerLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <OrganizerNavbar />
      <main className="flex flex-1 flex-col">{children}</main>
    </>
  );
}
