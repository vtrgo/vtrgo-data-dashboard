import { useFloatRange } from "@/hooks/useFloatRange";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { formatKey } from "@/utils/textFormat";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { formatDateTime } from "@/utils/timeFormat";

interface FloatAreaChartPanelProps {
  field: string;
  start: string;
  stop: string;
}

export function FloatAreaChartPanel({ field, start, stop }: FloatAreaChartPanelProps) {
  const { data, loading, error } = useFloatRange(field, { start, stop });

  if (loading) {
    return (
      <Card className="col-span-full">
        <CardHeader>
          <CardTitle>{formatKey(field)} - Historical Data</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">Loading historical data...</p>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className="col-span-full">
        <CardHeader>
          <CardTitle>{formatKey(field)} - Historical Data</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-64 text-red-500">
          <p>Error: {error.message}</p>
        </CardContent>
      </Card>
    );
  }

  if (!data || data.length === 0) {
    return (
      <Card className="col-span-full">
        <CardHeader>
          <CardTitle>{formatKey(field)} - Historical Data</CardTitle>
        </CardHeader>
        <CardContent className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">No data available for this range.</p>
        </CardContent>
      </Card>
    );
  }

  // Recharts expects data points with a time property that can be formatted
  // or a Date object. Convert the ISO string to Date objects.
  const chartData = data.map((d) => ({
    time: new Date(d.time), // Convert ISO string to Date object
    value: d.value,
  }));

  return (
    <Card className="col-span-full">
      <CardHeader>
        <CardTitle className="text-lg font-medium tracking-tight">
          {formatKey(field)} - Historical Data ({start} to {stop})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" className="stroke-gray-200" />
            <XAxis dataKey="time" tickFormatter={(tick) => formatDateTime(tick, "HH:mm")} minTickGap={30} angle={-45} textAnchor="end" height={70} />
            <YAxis />
            <Tooltip labelFormatter={(label) => formatDateTime(label, "MMM dd, HH:mm:ss")} formatter={(value: number) => [value.toFixed(2), "Value"]} />
            <Area type="monotone" dataKey="value" stroke="#8884d8" fill="#8884d8" fillOpacity={0.3} />
          </AreaChart>
        </ResponsiveContainer>
      </CardContent>
    </Card>
  );
}