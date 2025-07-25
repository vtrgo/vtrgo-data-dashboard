import { Skeleton } from "@/components/ui/skeleton";
import { DashboardLayout } from "./DashboardLayout";
import { DashboardSection } from "./DashboardSection";

export function DashboardSkeleton() {
  return (
    <DashboardLayout>
      {/* Skeleton for ProjectMetaPanel */}
      <DashboardSection>
        <Skeleton className="h-36 w-full" />
      </DashboardSection>

      <DashboardSection>
        <Skeleton className="h-48 w-full" />
      </DashboardSection>

      {/* Skeleton for HealthSummaryPanel */}
      <DashboardSection>
        <Skeleton className="h-48 w-full" />
      </DashboardSection>

      <DashboardSection>
        <div className="space-y-10">
          {/* Skeleton for FaultBarChartPanel */}
          <Skeleton className="h-[300px] w-full" />
          {/* Skeleton for FloatAreaChartPanel */}
          <Skeleton className="h-[300px] w-full" />
        </div>
      </DashboardSection>

      <DashboardSection>
        <div className="space-y-10">
          {/* Create a few skeleton panels */}
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-48 w-full" />
          ))}
        </div>
      </DashboardSection>

      <DashboardSection>
        <div className="space-y-10">
          {Array.from({ length: 2 }).map((_, i) => (
            <Skeleton key={i} className="h-48 w-full" />
          ))}
        </div>
      </DashboardSection>
    </DashboardLayout>
  );
}