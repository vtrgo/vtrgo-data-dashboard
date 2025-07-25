import React from "react";
import { cn } from "@/lib/utils";

interface DashboardSectionProps extends React.HTMLAttributes<HTMLElement> {
  children: React.ReactNode;
}

export function DashboardSection({ children, className, ...props }: DashboardSectionProps) {
  return (
    <section className={cn("font-serif", className)} {...props}>
      <div className="w-full max-w-2xl mx-auto">{children}</div>
    </section>
  );
}