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