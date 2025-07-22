// file: console/src/components/panels/FaultPanel.tsx
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getLabel } from "@/utils/databaseFields";
import { describeTimeRange } from "@/utils/describeTimeRange";

type FaultPanelProps = {
  faults: Record<string, number>;
  start: string;
  stop: string;
  className?: string;
};

export function FaultPanel({ faults, start, stop, className }: FaultPanelProps) {
  const entries = Object.entries(faults)
    .filter(([k]) => k.startsWith("FaultBits."))
    .sort(([, a], [, b]) => b - a)

  const { label } = describeTimeRange(start, stop);
  const title = `Fault Counts (Last ${label})`;

  return (
    <Card className={className}>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {entries.length === 0 ? (
          <p className="text-muted-foreground">No fault data available for this range.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <tbody>
                {entries.map(([key, value]) => (
                  <tr key={key} className="border-b last:border-b-0">
                    <td className="py-2 pr-4 font-medium">{getLabel(key)}</td>
                    <td className="py-2 text-right font-mono">{value}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
