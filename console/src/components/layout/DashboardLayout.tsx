import React from "react";

export function DashboardLayout({ children }: { children: React.ReactNode }) {
  return <main className="p-6 space-y-10">{children}</main>;
}