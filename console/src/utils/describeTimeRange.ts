import { durationHr } from "./timeFormat";

export function describeTimeRange(start: string, stop: string): { label: string; duration: string } {
  const now = new Date();

  const parseRelative = (val: string): Date => {
    if (val === "now()" || val === "now") return now;

    const match = val.match(/^-(\d+)([smhdw]|mo)$/); // supports s, m, h, d, w, mo
    if (!match) return new Date(val);

    const [, amountStr, unit] = match;
    const amount = parseInt(amountStr);
    const result = new Date(now);

    switch (unit) {
      case "s":
        result.setSeconds(now.getSeconds() - amount);
        break;
      case "m":
        result.setMinutes(now.getMinutes() - amount);
        break;
      case "h":
        result.setHours(now.getHours() - amount);
        break;
      case "d":
        result.setDate(now.getDate() - amount);
        break;
      case "w":
        result.setDate(now.getDate() - amount * 7);
        break;
      case "mo":
        result.setMonth(now.getMonth() - amount);
        break;
    }

    return result;
  };

  const s = parseRelative(start);
  const e = parseRelative(stop);
  const sStr = s.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
  const eStr = e.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });

  const durationMs = e.getTime() - s.getTime();
  const formattedDuration = durationHr(durationMs);

  let label = "";

  if (stop === "now()") {
    const match = start.match(/^-(\d+)([smhdw]|mo)$/);
    if (match) {
      const [, n, unit] = match;
      const unitMap: Record<string, string> = {
        s: "second",
        m: "minute",
        h: "hour",
        d: "day",
        w: "week",
        mo: "month",
      };
      const u = unitMap[unit] ?? unit;
      label = `Past ${n} ${parseInt(n) > 1 ? u + "s" : u}`;
    } else {
      label = `Up to ${eStr}`;
    }
  } else if (sStr === eStr) {
    label = sStr;
  } else {
    label = `${sStr} to ${eStr}`;
  }

  return {
    label,
    duration: formattedDuration,
  };
}
