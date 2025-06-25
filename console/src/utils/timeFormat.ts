export function durationHr(ms: number): string {
  const h = Math.floor(ms / 3600000);
  const m = Math.floor((ms % 3600000) / 60000);
  return `${h}h ${m}m`;
}

/**
 * Formats a Date object into a string based on a simple format string.
 * This is a basic implementation and does not support all date-fns formats.
 * It also defensively converts input to a Date object if it's not already one.
 */
export function formatDateTime(date: Date, formatStr: string): string {
  // Ensure 'd' is a Date object. If it's a string or number, try to convert it.
  let d: Date;
  if (date instanceof Date) {
    d = date;
  } else if (typeof date === 'string' || typeof date === 'number') {
    d = new Date(date);
    // Check if the conversion resulted in a valid Date object
    if (isNaN(d.getTime())) {
      // If not a valid date, return original value as string or a fallback
      return String(date);
    }
  } else {
    // If it's neither Date, string, nor number, return original value as string or a fallback
    return String(date);
  }

  const hours = d.getHours().toString().padStart(2, '0');
  const minutes = d.getMinutes().toString().padStart(2, '0');
  const seconds = d.getSeconds().toString().padStart(2, '0');
  const day = d.getDate().toString().padStart(2, '0');
  const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
  const month = monthNames[d.getMonth()];

  if (formatStr === "HH:mm") {
    return `${hours}:${minutes}`;
  } else if (formatStr === "MMM dd, HH:mm:ss") {
    return `${month} ${day}, ${hours}:${minutes}:${seconds}`;
  }
  return d.toLocaleString(); // Fallback for unsupported formats
}
