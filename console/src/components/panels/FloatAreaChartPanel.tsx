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

export function FloatAreaChartPanel({ field, start, stop, intervalMs }: FloatAreaChartPanelProps) {
  const { data, loading, error } = useFloatRange(field, { start, stop }, intervalMs);
  const title = `${formatKey(field)}`;
  const fieldKey = `${extractFieldKey(field)}`;
  const fieldUnit = getFieldUnit(fieldKey);

  // Common Card props for layout
  const cardProps = { className: 'col-span-full' };

  if (loading) {
    return (
      <Card {...cardProps}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">Loading historical data...</p>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card {...cardProps}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-64 text-destructive">
          <p>Error: {error.message}</p>
        </CardContent>
      </Card>
    );
  }

  if (!data || data.length === 0) {
    return (
      <Card {...cardProps}>
        <CardHeader>
          <CardTitle>{title}</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">No data available for this range.</p>
        </CardContent>
      </Card>
    );
  }

  // Format data for Recharts
  const chartData = data.map((d) => ({ time: new Date(d.time), value: d.value }));

  return (
    <Card {...cardProps}>
      <CardHeader>
        <CardTitle className="text-lg font-medium tracking-tight">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-neutral-200" />
            <XAxis
              dataKey="time"
              tickFormatter={(tick) => formatDateTime(tick, 'HH:mm')}
              minTickGap={30}
              angle={-45}
              textAnchor="end"
              height={70}
            />
            <YAxis />
            <Tooltip
              labelFormatter={(label) => formatDateTime(label, 'MMM dd, HH:mm:ss')}
              formatter={(value: number) => [`${value.toFixed(2)}  (${fieldUnit})`, `${fieldKey}`]}
            />
            <Area type="monotone" dataKey="value" stroke="var(--color-primary)" fill="var(--color-primary)" fillOpacity={0.3} />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}
