import { PanelGrid } from "@/components/layout/PanelGrid";
import { Skeleton } from "@/components/ui/skeleton";

export function DashboardSkeleton() {
  return (
    <main className="p-6 space-y-10">
      {/* Skeleton for ProjectMetaPanel */}
      <section className="font-serif">
        <Skeleton className="h-36 w-full max-w-2xl mx-auto" />
      </section>

      <section className="font-serif">
        <Skeleton className="h-48 w-full max-w-2xl mx-auto" />
      </section>

      <section className="font-serif">
        {/* Skeleton for HealthSummaryPanel */}
        <Skeleton className="h-48 w-full max-w-2xl mx-auto" />
      </section>

      <section className="font-serif">
        <div className="space-y-10">
          {/* Skeleton for FaultBarChartPanel */}
          <Skeleton className="h-[300px] w-full max-w-2xl mx-auto" />
          {/* Skeleton for FloatAreaChartPanel */}
          <Skeleton className="h-[300px] w-full max-w-2xl mx-auto" />
        </div>
      </section>

      <section className="font-serif">
        <PanelGrid>
          {/* Create a few skeleton panels */}
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-48 w-full" />
          ))}
        </PanelGrid>
      </section>

      <section className="font-serif">
        <PanelGrid>
          {Array.from({ length: 2 }).map((_, i) => (
            <Skeleton key={i} className="h-48 w-full" />
          ))}
        </PanelGrid>
      </section>
    </main>
  );
}