import type { ReactNode } from "react";

export function PanelGrid({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-muted grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 p-4">
      {children}
    </div>
  );
}
