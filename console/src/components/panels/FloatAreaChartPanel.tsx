// file: console/src/components/panels/FloatAreaChartPanel.tsx
import { useFloatRange } from '@/hooks/useFloatRange';
import { Card, CardHeader, CardContent, CardTitle } from '@/components/ui/card';
import { formatKey, extractFieldKey, getFieldUnit } from '@/utils/textFormat';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { formatDateTime } from '@/utils/timeFormat';

interface FloatAreaChartPanelProps {
  field: string;
  start: string;
  stop: string;
  intervalMs?: number | null;
}

/**
 * A panel that displays a time-series float value as an area chart.
 * Styling is consistent with shadcn/ui theming principles, using CSS variables.
 */
export function FloatAreaChartPanel({ field, start, stop, intervalMs }: FloatAreaChartPanelProps) {
  const { data, loading, error } = useFloatRange(field, { start, stop }, intervalMs);
  const title = `${formatKey(field)}`;
  const fieldKey = `${extractFieldKey(field)}`;
  const fieldUnit = getFieldUnit(fieldKey);

  const cardProps = { className: 'col-span-full' };

  // Loading State
  if (loading) {
    return (
      <Card {...cardProps}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex h-64 items-center justify-center">
          <p className="text-muted-foreground">Loading historical data...</p>
        </CardContent>
      </Card>
    );
  }

  // Error State
  if (error) {
    return (
      <Card {...cardProps}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex h-64 items-center justify-center text-destructive">
          <p>Error: {error.message}</p>
        </CardContent>
      </Card>
    );
  }

  // No Data State
  if (!data || data.length === 0) {
    return (
      <Card {...cardProps}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex h-64 items-center justify-center">
          <p className="text-muted-foreground">No data available for this range.</p>
        </CardContent>
      </Card>
    );
  }

  const chartData = data.map((d) => ({ time: new Date(d.time), value: d.value }));

  return (
    <Card {...cardProps}>
      <CardHeader>
        <CardTitle className="text-lg font-medium tracking-tight">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 10, bottom: 0 }}>
            <CartesianGrid stroke="var(--border)" strokeDasharray="3 3" />

            <XAxis
              dataKey="time"
              tickFormatter={(tick) => formatDateTime(tick, 'HH:mm')}
              tick={{ fill: 'var(--muted-foreground)', fontSize: 12 }}
              axisLine={{ stroke: 'var(--border)' }}
              tickLine={{ stroke: 'var(--border)' }}
              minTickGap={30}
              angle={-45}
              textAnchor="end"
              height={70}
            />

            <YAxis
              tick={{ fill: 'var(--muted-foreground)', fontSize: 12 }}
              axisLine={{ stroke: 'var(--border)' }}
              tickLine={{ stroke: 'var(--border)' }}
            />

            <Tooltip
              cursor={{ fill: 'var(--muted)' }}
              labelFormatter={(label) => formatDateTime(label, 'MMM dd, HH:mm:ss')}
              formatter={(value: number) => [`${value.toFixed(2)} (${fieldUnit})`, `${fieldKey}`]}
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
              <linearGradient id="colorPrimary" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.8} />
                <stop offset="95%" stopColor="var(--primary)" stopOpacity={0.1} />
              </linearGradient>
            </defs>

            <Area
              type="monotone"
              dataKey="value"
              stroke="var(--primary)"
              fill="url(#colorPrimary)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
