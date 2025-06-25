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

export function useFloatRange(field: string, { start = "-1h", stop = "now()" }: TimeRange = {}) {
  const [data, setData] = useState<FloatDataPoint[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    // Don't fetch if no field is provided
    if (!field) {
      setData(null);
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);
    fetch(`/api/float-range?field=${encodeURIComponent(field)}&start=${encodeURIComponent(start)}&stop=${encodeURIComponent(stop)}`)
      .then((res) => {
        if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
        return res.json();
      })
      .then(setData)
      .catch(setError)
      .finally(() => setLoading(false));
  }, [field, start, stop]);

  return { data, loading, error };
}