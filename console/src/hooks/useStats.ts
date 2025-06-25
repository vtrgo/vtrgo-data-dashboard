import { useEffect, useState } from "react";

type TimeRange = {
  start?: string;
  stop?: string;
};

// Added intervalMs for auto-refresh
export function useStats(
  { start = "-1h", stop = "now()" }: TimeRange = {},
  intervalMs: number | null = null // New parameter: refresh interval in milliseconds
) {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null); // Added error state for robustness

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      setError(null); // Clear previous errors
      try {
        const response = await fetch(`/api/stats?start=${encodeURIComponent(start)}&stop=${encodeURIComponent(stop)}`);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
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
  }, [start, stop, intervalMs]); // Re-run effect if start, stop, or intervalMs changes

  return { data, loading, error };
}
