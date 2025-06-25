import { useEffect, useState } from "react";

type TimeRange = {
  start?: string;
  stop?: string;
};

// Define the shape of a single data point for the chart
type FloatDataPoint = {
  time: string; // ISO 8601 date string
  value: number;
};

export function useFloatRange(
  field: string,
  { start = "-1h", stop = "now()" }: TimeRange = {},
  intervalMs: number | null = null // New parameter: refresh interval in milliseconds
) {
  const [data, setData] = useState<FloatDataPoint[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      // Don't fetch if no field is provided
      if (!field) {
        setData(null);
        setLoading(false);
        return;
      }

      setLoading(true);
      setError(null); // Clear previous errors
      try {
        const response = await fetch(`/api/float-range?field=${encodeURIComponent(field)}&start=${encodeURIComponent(start)}&stop=${encodeURIComponent(stop)}`);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const json = await response.json();
        setData(json);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };

    fetchData(); // Initial fetch when component mounts or dependencies change

    let intervalId: NodeJS.Timeout | undefined;
    if (intervalMs && intervalMs > 0) {
      intervalId = setInterval(fetchData, intervalMs); // Set up auto-refresh
    }

    return () => {
      // Cleanup function: clear interval when component unmounts or dependencies change
      if (intervalId) {
        clearInterval(intervalId);
      }
    };
  }, [field, start, stop, intervalMs]); // Re-run effect if field, start, stop, or intervalMs changes

  return { data, loading, error };
}