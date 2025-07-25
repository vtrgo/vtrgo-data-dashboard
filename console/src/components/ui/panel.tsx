import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ReactNode } from "react";

export interface PanelProps {
  title: string;
  children: ReactNode;
  className?: string;
}

/**
 * A generic wrapper component that provides a consistent Card-based UI for panels.
 */
export function Panel({ title, children, className }: PanelProps) {
  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle className="text-lg font-medium tracking-tight">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {children}
      </CardContent>
    </Card>
  );
}