// file: console/src/components/panels/FaultBarChartPanel.tsx

import { useMemo } from 'react';
import { Card, CardHeader, CardContent, CardTitle } from '@/components/ui/card';
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

interface FaultBarChartPanelProps {
  faults: Record<string, number>;
}

/**
 * A panel that displays fault counts in a horizontal bar chart.
 * It uses the centralized databaseFields utility for formatting keys.
 */
export function FaultBarChartPanel({ faults }: FaultBarChartPanelProps) {
  const title = 'Fault Counts';
  const cardProps = { className: 'col-span-full' };

  const chartData = useMemo(() => {
    if (!faults || Object.keys(faults).length === 0) {
      return [];
    }
    return Object.entries(faults)
      .filter(([_, count]) => count > 0)
      .map(([faultKey, count]) => ({
        fault: faultKey,
        label: getLabel(faultKey), // Use the new utility function
        count,
      }))
      .sort((a, b) => b.count - a.count);
  }, [faults]);

  const hasData = chartData.length > 0;

  // Dynamically calculate Y-axis width based on the longest label
  const yAxisWidth = useMemo(() => {
    if (!hasData) return 100;
    const longestLabel = chartData.reduce((max, item) => {
      return item.label.length > max ? item.label.length : max;
    }, 0);
    // Estimate width: 8px per character plus 40px padding
    return Math.max(120, longestLabel * 8 + 40);
  }, [chartData, hasData]);


  if (!hasData) {
    return (
      <Card {...cardProps}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex h-64 items-center justify-center">
          <p className="text-muted-foreground">No fault data available for this range.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card {...cardProps}>
      <CardHeader>
        <CardTitle className="text-lg font-medium tracking-tight">{title}</CardTitle>
      </CardHeader>
      <CardContent>
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
              cursor={{ fill: 'var(--muted)' }}
              formatter={(value: number) => [value, 'Count']}
              contentStyle={{
                background: 'var(--popover)',
                borderColor: 'var(--border)',
                borderRadius: 'var(--radius)',
                color: 'var(--popover-foreground)',
              }}
              labelStyle={{
                color: 'var(--popover-foreground)',
                fontWeight: 500,
              }}
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
              stroke="var(--primary)"
              radius={[0, 4, 4, 0]}
            >
              <LabelList
                dataKey="count"
                position="right"
                offset={8}
                style={{
                  fill: 'var(--primary-foreground)',
                  fontSize: 14,
                  fontWeight: 500,
                }}
              />
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
