import { Card, CardHeader, CardContent, CardTitle } from '@/components/ui/card';
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
 * Styling is consistent with shadcn/ui theming principles, using CSS variables.
 */
export function FaultBarChartPanel({ faults }: FaultBarChartPanelProps) {
  const title = 'Fault Counts';
  const cardProps = { className: 'col-span-full' };

  const hasData = faults && Object.keys(faults).length > 0;

  // Map faults object to array, remove "FaultBits." prefix for display, and sort
  const chartData = hasData
    ? Object.entries(faults)
        .filter(([_, count]) => count > 0)
        .map(([fault, count]) => ({
          fault,
          label: fault.startsWith('FaultBits.') ? fault.slice('FaultBits.'.length) : fault,
          count,
        }))
        .sort((a, b) => b.count - a.count)
    : [];

  // Dynamically calculate Y-axis width based on the longest label
  const yAxisWidth = hasData
    ? Math.max(
        100, // Minimum width
        chartData.reduce((maxWidth, item) => {
          // Estimate text width using the new label property and add padding
          const currentWidth = item.label.length * 8 + 40;
          return Math.max(maxWidth, currentWidth);
        }, 0)
      )
    : 100;

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
