import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { ReactNode } from "react";

interface ChartPanelProps {
  /** The title displayed in the card header. */
  title: ReactNode;
  /** Optional content to render next to the title in the header. */
  headerContent?: ReactNode;
  /** The main content, typically a chart, to display when data is available. */
  children: ReactNode;
  /** The loading state of the data. */
  isLoading: boolean;
  /** An error object if the data fetch failed. */
  error?: Error | null;
  /** A boolean to indicate if there is data to display. */
  hasData: boolean;
  /** Custom class name for the root Card element. */
  className?: string;
  /** The height of the content area, to prevent layout shifts during loading. */
  contentHeight?: string | number;
}

/**
 * A generic wrapper component for displaying charts. It provides a consistent
 * Card-based UI and handles loading, error, and empty states.
 */
export function ChartPanel({
  title,
  headerContent,
  children,
  isLoading,
  error,
  hasData,
  className,
  contentHeight = 300,
}: ChartPanelProps) {
  return (
    <Card className={className}>
      <CardHeader>
        <div className="flex justify-between items-start gap-4">
          <CardTitle className="text-lg font-medium tracking-tight">{title}</CardTitle>
          {headerContent}
        </div>
      </CardHeader>
      <CardContent style={{ height: contentHeight }}>
        {isLoading && (
          <div className="flex h-full items-center justify-center"><p className="text-muted-foreground">Loading...</p></div>
        )}
        {error && (
          <div className="flex h-full items-center justify-center text-destructive"><p>Error: {error.message}</p></div>
        )}
        {!isLoading && !error && !hasData && (
          <div className="flex h-full items-center justify-center"><p className="text-muted-foreground">No data available.</p></div>
        )}
        {!isLoading && !error && hasData && children}
      </CardContent>
    </Card>
  );
}