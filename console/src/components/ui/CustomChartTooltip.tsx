import { formatDateTime } from '@/utils/timeFormat';

/**
 * Defines the shape of the data item passed in the payload array by Recharts.
 */
interface PayloadItem {
  name: string;
  value: number;
  unit?: string;
  stroke?: string;
  color?: string;
  payload?: {
    label?: string;
  };
}

/**
 * Defines the props passed to the custom tooltip component by Recharts.
 */
interface CustomTooltipProps {
  active?: boolean;
  payload?: PayloadItem[];
  label?: string | number;
}

/**
 * A shared, reusable tooltip component for all charts to ensure consistent styling.
 * It uses shadcn/ui CSS variables for theming.
 */
export const CustomChartTooltip = ({ active, payload, label }: CustomTooltipProps) => {
  // The `label` can be undefined if the tooltip is not active over a data point.
  if (active && payload && payload.length && typeof label !== 'undefined') {
    // The 'label' from recharts can be a timestamp (number) or a string.
    // We convert it to a Date object to format it reliably.
    const date = new Date(label);
    const formattedLabel = !isNaN(date.getTime()) // Check if the date is valid before formatting
      ? formatDateTime(date, 'MMM dd, HH:mm:ss')
      : String(label);

    return (
      <div className="rounded-lg border bg-popover p-2 text-sm shadow-sm animate-in fade-in-0 zoom-in-95">
        <div className="font-medium text-popover-foreground">{formattedLabel}</div>
        {payload.map((p, i) => (
          <div key={i} className="text-muted-foreground">
            {/* Use p.name for the series label, which is set on the <Bar /> or <Area /> component. */}
            {p.name}:{' '}
            <span
              className="font-bold"
              // Prioritize `stroke` for color, as `color` can be a gradient URL for bars.
              style={{ color: p.stroke || p.color || 'hsl(var(--primary))' }}
            >
              {p.value?.toLocaleString()}
              {p.unit ? ` ${p.unit}` : ''}
            </span>
          </div>
        ))}
      </div>
    );
  }

  return null;
};
