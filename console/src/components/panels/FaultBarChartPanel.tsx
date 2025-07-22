// file: console/src/components/panels/FaultBarChartPanel.tsx

import { useMemo } from 'react';
import { ChartPanel } from '@/components/panels/ChartPanel';
import { getLabel } from '@/utils/databaseFields';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  LabelList,
} from 'recharts';
import { CustomChartTooltip } from '@/components/charts/CustomChartTooltip';

interface FaultBarChartPanelProps {
  faults: Record<string, number>;
  /** The maximum number of faults to display in the chart. Defaults to 15. */
  maxItems?: number;
  className?: string;
}

/**
 * A panel that displays fault counts in a horizontal bar chart.
 * It uses the centralized `getLabel` utility for formatting keys.
 */
export function FaultBarChartPanel({
  faults,
  maxItems = 15,
  className,
}: FaultBarChartPanelProps) {
  const title = 'Fault Counts';

  const allFaults = useMemo(() => {
    if (!faults || Object.keys(faults).length === 0) {
      return [];
    }
    return Object.entries(faults)
      .filter(([_, count]) => count > 0)
      .map(([faultKey, count]) => ({
        fault: faultKey,
        label: getLabel(faultKey),
        count,
      }))
      .sort((a, b) => b.count - a.count);
  }, [faults]);

  const chartData = useMemo(() => allFaults.slice(0, maxItems), [allFaults, maxItems]);

  const hasData = chartData.length > 0;
  const totalFaults = allFaults.length;
  const isTruncated = totalFaults > maxItems;

  // Dynamically calculate Y-axis width based on the longest label
  const yAxisWidth = useMemo(
    () => {
      if (!hasData) return 100;
      const longestLabel = chartData.reduce((max, item) => (item.label.length > max ? item.label.length : max), 0);
      // Estimate width: ~7px per char + 40px padding. Cap at 300px to prevent layout breakage.
      return Math.min(300, Math.max(120, longestLabel * 7 + 40));
    },
    [chartData, hasData]
  );

  return (
    <ChartPanel title={title} isLoading={false} hasData={hasData} className={className} contentHeight={chartData.length * 40 + 60}>
      <div className="flex flex-col h-full">
        <ResponsiveContainer width="100%" height={chartData.length * 40 + 30}>
          <BarChart
            data={chartData}
            layout="vertical"
            margin={{ top: 5, right: 50, left: 10, bottom: 5 }}
            barCategoryGap="25%"
          >
            <XAxis
              type="number"
              axisLine={false}
              tickLine={false}
              tick={{ fill: 'var(--muted-foreground)', fontSize: 12 }}
              allowDecimals={false}
            />

            <YAxis
              dataKey="label"
              type="category"
              width={yAxisWidth}
              tickLine={false}
              axisLine={false}
              tick={{
                fill: 'var(--foreground)',
                fontWeight: 500,
                fontSize: 14,
              }}
            />

            <Tooltip
              cursor={{ fill: 'hsl(var(--muted))' }}
              content={<CustomChartTooltip />}
            />

            <defs>
              <linearGradient id="barFillGradient" x1="0" y1="0" x2="1" y2="0">
                <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.1} />
                <stop offset="95%" stopColor="var(--primary)" stopOpacity={0.6} />
              </linearGradient>
            </defs>

            <Bar
              dataKey="count"
              fill="url(#barFillGradient)"
              name="Count"
              stroke="var(--primary)"
              radius={[0, 4, 4, 0]}
            >
              <LabelList
                dataKey="count"
                position="right"
                offset={8}
                className="fill-foreground text-sm font-medium"
              />
            </Bar>
          </BarChart>
        </ResponsiveContainer>
        {isTruncated && (
          <p className="text-center text-xs text-muted-foreground mt-2">
            Showing top {maxItems} of {totalFaults} faults.
          </p>
        )}
      </div>
    </ChartPanel>
  );
}
