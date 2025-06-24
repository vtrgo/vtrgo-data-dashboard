import { useEffect, useState } from "react";

type TimeRange = {
  start?: string;
  stop?: string;
};

export function useStats({ start = "-1h", stop = "now()" }: TimeRange = {}) {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    fetch(`/api/stats?start=${encodeURIComponent(start)}&stop=${encodeURIComponent(stop)}`)
      .then((res) => res.json())
      .then(setData)
      .finally(() => setLoading(false));
  }, [start, stop]);

  return { data, loading };
}
