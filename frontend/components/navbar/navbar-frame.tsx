import React from "react";
import { cn } from "@/lib/utils";

export interface NavbarFrameProps {
  left?: React.ReactNode;
  right?: React.ReactNode;
  children?: React.ReactNode;
  className?: string;
}

export function NavbarFrame({
  left,
  right,
  children,
  className,
}: NavbarFrameProps) {
  return (
    <header
      className={cn(
        "bg-background/95 sticky top-0 z-50 w-full border-b backdrop-blur",
        className
      )}
    >
      <div className="container mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-8">
        {left}
        {right}
        {children}
      </div>
    </header>
  );
}
