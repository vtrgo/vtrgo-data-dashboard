// file: console/src/components/panels/FloatAreaChartPanel.tsx

import { useState, useEffect, useMemo } from 'react';
import { useFloatRange } from '@/hooks/useFloatRange';
import { ChartPanel } from '@/components/panels/ChartPanel';
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
import { CustomChartTooltip } from '@/components/charts/CustomChartTooltip';

interface FloatAreaChartPanelProps {
  floatFields: string[];
  start: string;
  stop: string;
  intervalMs?: number | null;
  className?: string;
}

/**
 * A panel that displays a time-series float value as an area chart, with field selection.
 * It uses the centralized databaseFields utility for parsing and formatting keys.
 */
export function FloatAreaChartPanel({
  floatFields,
  start,
  stop,
  intervalMs,
  className,
}: FloatAreaChartPanelProps) {
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

  const chartData = useMemo(() => {
    if (!data) return [];
    return data.map((d) => ({ time: new Date(d.time), value: d.value }));
  }, [data]);

  const tickFormatter = useMemo(() => {
    // A simple way to check if the range is more than a day.
    // A more robust solution might parse the start string more carefully.
    const isMultiDay = start.includes('d') || start.includes('w') || start.includes('mo');
    return (tickValue: string | number | Date) => {
      // Recharts can pass a timestamp (number) or a date string for the ticks,
      // so we ensure it's always a Date object before formatting.
      const date = new Date(tickValue);
      if (isNaN(date.getTime())) {
        // If the date is invalid, return an empty string to avoid display errors.
        return '';
      }
      return isMultiDay ? `${date.getMonth() + 1}/${date.getDate()}` : `${date.getHours()}:${String(date.getMinutes()).padStart(2, '0')}`;
    };
  }, [start]);

  const headerContent = (
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
  );

  return (
    <ChartPanel
      title={`${groupLabel} Data`}
      headerContent={headerContent}
      isLoading={loading}
      error={error}
      hasData={chartData.length > 0}
      className={className}
    >
      <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 10, bottom: 0 }}>
                <CartesianGrid stroke="var(--border)" strokeDasharray="3 3" />
                <XAxis
                  dataKey="time"
                  tickFormatter={tickFormatter}
                  tick={{ fill: 'var(--muted-foreground)', fontSize: 12 }}
                  axisLine={{ stroke: 'var(--border)' }}
                  tickLine={{ stroke: 'var(--border)' }}
                  minTickGap={40}
                  angle={-45}
                  textAnchor="end"
                  height={70}
                />
                <YAxis
                  tick={{ fill: 'var(--muted-foreground)', fontSize: 12 }}
                  axisLine={{ stroke: 'var(--border)' }}
                  tickLine={{ stroke: 'var(--border)' }}
                />
                <Tooltip cursor={{ fill: 'hsl(var(--muted))' }} content={<CustomChartTooltip />} />
                <defs>
                  <linearGradient id="colorPrimary" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="var(--primary)" stopOpacity={0.8} />
                    <stop offset="95%" stopColor="var(--primary)" stopOpacity={0.1} />
                  </linearGradient>
                </defs>
                <Area
                  type="monotone"
                  dataKey="value"
                  unit={fieldUnit}
                  name={formatSegment(leafKey)}
                  stroke="var(--primary)"
                  fill="url(#colorPrimary)"
                  dot={false}
                />
              </AreaChart>
      </ResponsiveContainer>
    </ChartPanel>
  );
}
