// Converts a JS Date or string to RFC3339 format for API compatibility
export function toApiTimeValue(value: Date | string): string {
  if (typeof value === 'string') {
    // Pass through relative time strings like '-1h', 'now()'
    if (/^(-?\d+[smhdw]|now\(\))$/.test(value)) return value;
    // Try to parse as date string
    const date = new Date(value);
    if (!isNaN(date.getTime())) {
      return date.toISOString(); // RFC3339
    }
    return value; // fallback
  }
  // JS Date object
  return value.toISOString();
}
