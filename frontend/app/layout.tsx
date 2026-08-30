import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { cn } from "@/lib/utils";
import { RoleProvider } from "@/context/role-context";
import { Navbar } from "@/components/navbar";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-sans",
});

export const metadata: Metadata = {
  title: "Eventory - Competition & Tournament Platform",
  description:
    "Manage, compete, and sponsor tournaments with unified role access.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={cn("h-full font-sans antialiased", inter.variable)}
    >
      <body className="bg-background text-foreground flex min-h-full flex-col">
        <RoleProvider>
          <Navbar />
          <main className="flex flex-1 flex-col">{children}</main>
        </RoleProvider>
      </body>
    </html>
  );
}
