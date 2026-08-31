import React from "react";
import { Navbar } from "@/components/navbar";

export default function PublicLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <>
      <Navbar />
      <main className="flex flex-1 flex-col">{children}</main>
    </>
  );
}
