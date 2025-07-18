// file: console/src/components/panels/FloatAreaChartPanel.tsx

import { useState, useEffect, useMemo } from 'react';
import { useFloatRange } from '@/hooks/useFloatRange';
import { Card, CardHeader, CardContent, CardTitle } from '@/components/ui/card';
import {
  getUnit,
  getGroupLabel,
  getFieldLabel,
  parseKey,
  formatSegment,
} from '@/utils/databaseFields';
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
  floatFields: string[];
  start: string;
  stop: string;
  intervalMs?: number | null;
}

/**
 * A panel that displays a time-series float value as an area chart, with field selection.
 * It uses the centralized databaseFields utility for parsing and formatting keys.
 */
export function FloatAreaChartPanel({ floatFields, start, stop, intervalMs }: FloatAreaChartPanelProps) {
  const [selectedField, setSelectedField] = useState<string>(floatFields[0] || '');

  useEffect(() => {
    // Reset selectedField if it's no longer in the list of available fields
    if (floatFields.length > 0 && !floatFields.includes(selectedField)) {
      setSelectedField(floatFields[0]);
    } else if (floatFields.length === 0) {
      setSelectedField('');
    }
  }, [floatFields, selectedField]);

  const { data, loading, error } = useFloatRange(selectedField, { start, stop }, intervalMs);

  // Derive all display values from the selected field key using the new utils
  const parsedKey = useMemo(() => (selectedField ? parseKey(selectedField) : null), [selectedField]);
  const groupLabel = useMemo(() => (parsedKey ? getGroupLabel(parsedKey) : 'Performance Data'), [parsedKey]);
  const fieldUnit = useMemo(() => (parsedKey ? getUnit(parsedKey) : ''), [parsedKey]);
  const leafKey = useMemo(() => (parsedKey ? parsedKey.subgroup || parsedKey.field : ''), [parsedKey]);

  return (
    <Card className="col-span-full">
      <CardHeader>
        <div className="flex justify-between items-start gap-4">
          {/* Title on the left */}
          <CardTitle className="text-lg font-medium tracking-tight">
            {groupLabel} Data
          </CardTitle>

          {/* Selector on the right */}
          <select
            className="border px-2 py-1 text-sm bg-background rounded-md shadow-sm hover:border-primary focus:outline-none focus:ring-2 focus:ring-ring w-fit"
            onChange={(e) => setSelectedField(e.target.value)}
            value={selectedField}
            disabled={floatFields.length <= 1}
            aria-label="Select a data field to display"
          >
            {floatFields.length === 0 && <option>No fields available</option>}
            {floatFields.map((fieldKey) => (
              <option key={fieldKey} value={fieldKey}>
                {getFieldLabel(fieldKey)}
              </option>
            ))}
          </select>
        </div>
      </CardHeader>
      <CardContent>
        {loading && (
          <div className="flex h-64 items-center justify-center">
            <p className="text-muted-foreground">Loading historical data...</p>
          </div>
        )}
        {error && (
          <div className="flex h-64 items-center justify-center text-destructive">
            <p>Error: {error.message}</p>
          </div>
        )}
        {!loading && !error && (!data || data.length === 0) && (
          <div className="flex h-64 items-center justify-center">
            <p className="text-muted-foreground">No data available for this range.</p>
          </div>
        )}
        {!loading && !error && data && data.length > 0 && (
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={data.map((d) => ({ time: new Date(d.time), value: d.value }))} margin={{ top: 10, right: 30, left: 10, bottom: 0 }}>
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
                formatter={(value: number) => [
                  `${value.toFixed(2)} ${fieldUnit ? `(${fieldUnit})` : ''}`,
                  formatSegment(leafKey),
                ]}
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
                dot={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        )}
      </CardContent>
    </Card>
  );
}
