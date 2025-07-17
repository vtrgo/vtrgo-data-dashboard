// file: console/src/utils/textFormat.ts
export function formatFaultKey(key: string): string {
  return key.replace("FaultBits.", "").replace(/_/g, " ");
}

export function formatKey(key: string): string {
  // Add space before capital letters in a camelCase word, but not for acronyms
  let formattedKey = key.replace(/([a-z])([A-Z])/g, "$1 $2");

  // Handle module prefixes like "BMS."
  formattedKey = formattedKey.replace(/\./g, " - ");

  // Replace underscores with spaces
  formattedKey = formattedKey.replace(/_/g, " ");

  return formattedKey;
}

export function extractFieldKey(key: string): string {
  const parts = key.split('.');
  return parts[parts.length - 1] || key;
}

export function getFieldUnit(fieldKey: string): string {
  if (/^Vibration[XYZ]$/.test(fieldKey)) {
    return "mm/s²";
  }

  switch (fieldKey) {
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
