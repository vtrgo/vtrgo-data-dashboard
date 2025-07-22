import { PanelGrid } from "@/components/layout/PanelGrid";
import { Skeleton } from "@/components/ui/skeleton";

export function DashboardSkeleton() {
  return (
    <>
      {/* Skeleton for Performance Data section */}
      <section className="font-serif">
        <h2 className="pl-9 pt-3 text-xl uppercase tracking-widest text-muted-foreground mb-4 italic">
          <Skeleton className="h-6 w-64" />
        </h2>
        <PanelGrid>
          {/* Skeleton for FloatAreaChartPanel and FaultBarChartPanel */}
          <Skeleton className="h-[300px] w-full col-span-1 md:col-span-2" />
          <Skeleton className="h-[300px] w-full col-span-1" />
        </PanelGrid>
      </section>

      {/* Skeleton for main content sections */}
      <main className="p-6 space-y-10">
        <section className="font-serif">
          <h2 className="text-xl uppercase tracking-widest text-muted-foreground mb-4 italic">
            <Skeleton className="h-6 w-96" />
          </h2>
          <PanelGrid>
            {/* Create a few skeleton panels */}
            {Array.from({ length: 3 }).map((_, i) => (
              <Skeleton key={i} className="h-48 w-full" />
            ))}
          </PanelGrid>
        </section>

        <section className="font-serif">
          <h2 className="text-xl uppercase tracking-widest text-muted-foreground mb-4 italic">
            <Skeleton className="h-6 w-48" />
          </h2>
          <PanelGrid>
            {Array.from({ length: 2 }).map((_, i) => (
              <Skeleton key={i} className="h-48 w-full" />
            ))}
          </PanelGrid>
        </section>
      </main>
    </>
  );
}