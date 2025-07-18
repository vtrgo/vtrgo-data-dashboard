// file: console/src/utils/databaseFields.ts

/**
 * Represents the parsed structure of a database field key.
 * The expected format is `group.field.subgroup`.
 */
export interface ParsedKey {
  /** The original, raw key. e.g., "Floats.AirTrackBlower.Speed" */
  readonly raw: string;
  /** The top-level grouping. e.g., "Floats" */
  readonly group: string | null;
  /** The main field identifier. e.g., "AirTrackBlower" */
  readonly field: string;
  /** The specific sub-part of the field. e.g., "Speed" */
  readonly subgroup: string | null;
}

/**
 * Parses a raw database field key into its constituent parts.
 * @param key The raw string key from the database.
 * @returns A `ParsedKey` object.
 */
export function parseKey(key: string): ParsedKey {
  const parts = key.split('.');
  switch (parts.length) {
    case 1:
      // Key has no group, e.g., "Temperature"
      return { raw: key, group: null, field: parts[0], subgroup: null };
    case 2:
      // Key has group and field, e.g., "SystemStatusBits.AirPressureOk"
      return { raw: key, group: parts[0], field: parts[1], subgroup: null };
    default:
      // Key has group, field, and subgroup(s), e.g., "FaultBits.JamInOrientation.Lane1"
      // All parts after the second are joined to form the subgroup.
      return {
        raw: key,
        group: parts[0],
        field: parts[1],
        subgroup: parts.slice(2).join('.'),
      };
  }
}

/**
 * Formats a single string segment by splitting camelCase and replacing underscores.
 * @param segment The string segment to format.
 * @returns A formatted, human-readable string.
 */
export function formatSegment(segment: string): string {
  if (!segment) return "";
  // Add space before capital letters in a camelCase word, then replace underscores
  return segment
    .replace(/([a-z])([A-Z])/g, "$1 $2")
    .replace(/_/g, " ");
}

/**
 * Generates a full, human-readable label from a key.
 * It intelligently omits common prefixes like "FaultBits" for clarity.
 * @param key A `ParsedKey` object or a raw string key.
 * @returns A formatted string label for UI display. e.g., "Air Track Blower - Speed"
 */
export function getLabel(key: string | ParsedKey): string {
  const parsed = typeof key === 'string' ? parseKey(key) : key;
  const { group, field, subgroup } = parsed;

  // Omit common, non-descriptive group prefixes from the final label.
  const SKIPPED_GROUPS = ["FaultBits", "SystemStatusBits", "StatusBits"];
  const displayGroup = group && !SKIPPED_GROUPS.includes(group)
    ? formatSegment(group)
    : null;

  const displayField = formatSegment(field);
  const displaySubgroup = subgroup ? formatSegment(subgroup) : null;

  return [displayGroup, displayField, displaySubgroup].filter(Boolean).join(' - ');
}

/**
 * Generates a human-readable label for the group part of a key.
 * @param key A ParsedKey object or a raw string key.
 * @returns The formatted group name. e.g., "Floats"
 */
export function getGroupLabel(key: string | ParsedKey): string {
  const parsed = typeof key === 'string' ? parseKey(key) : key;
  return parsed.group ? formatSegment(parsed.group) : '';
}

/**
 * Generates a human-readable label for the field and subgroup parts of a key.
 * @param key A ParsedKey object or a raw string key.
 * @returns The formatted field and subgroup, joined together. e.g., "Air Track Blower - Speed"
 */
export function getFieldLabel(key: string | ParsedKey): string {
  const parsed = typeof key === 'string' ? parseKey(key) : key;
  const displayField = formatSegment(parsed.field);
  const displaySubgroup = parsed.subgroup ? formatSegment(parsed.subgroup) : null;
  return [displayField, displaySubgroup].filter(Boolean).join(' - ');
}

/**
 * Determines the appropriate unit for a given database key.
 * The unit is determined by the most specific part of the key (subgroup or field).
 * @param key A `ParsedKey` object or a raw string key.
 * @returns The unit symbol (e.g., "°C", "Hz") or an empty string.
 */
export function getUnit(key: string | ParsedKey): string {
  const parsed = typeof key === 'string' ? parseKey(key) : key;
  const leafKey = parsed.subgroup || parsed.field;

  if (/^Vibration[XYZ]$/.test(leafKey)) {
    return "mm/s²";
  }

  switch (leafKey) {
    case "Temperature":
      return "°C";
    case "Speed":
      return "Hz";
    case "PartsPerMinute":
      return "PPM";
    default:
      return "";
  }
}
