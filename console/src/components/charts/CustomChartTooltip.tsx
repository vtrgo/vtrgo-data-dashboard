/**
 * Defines the shape of the data item passed in the payload array.
 */
interface PayloadItem {
  name: string;
  value: number;
  color?: string;
}

/**
 * Defines the props passed to the custom tooltip component by Recharts.
 */
interface CustomTooltipProps {
  active?: boolean;
  payload?: PayloadItem[];
  label?: string | number | Date; // Label can be various types
}

/**
 * A shared, reusable tooltip component for all charts to ensure consistent styling.
 * It uses shadcn/ui CSS variables for theming.
 */
export const CustomChartTooltip = ({ active, payload, label }: CustomTooltipProps) => {
  if (active && payload && payload.length) {
    // Format the label appropriately before rendering
    const formattedLabel =
      label instanceof Date
        ? label.toLocaleString([], {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          })
        : label;

    return (
      <div className="rounded-lg border bg-popover p-2 text-sm shadow-sm">
        <div className="font-medium text-popover-foreground">{formattedLabel}</div>
        {payload.map((p: PayloadItem, i: number) => (
          <div key={i} className="text-muted-foreground">
            {p.name}:{' '}
            <span className="font-bold" style={{ color: p.color || 'var(--primary)' }}>
              {p.value}
            </span>
          </div>
        ))}
      </div>
    );
  }

  return null;
};
