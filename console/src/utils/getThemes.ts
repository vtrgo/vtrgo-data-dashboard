// console/src/utils/getThemes.ts

/**
 * Scans all loaded stylesheets for `.theme-XYZ` class selectors
 * Returns a de-duplicated list of theme names
 */
export function getAvailableThemes(): string[] {
  const seen = new Set<string>();
  const results: string[] = [];

  for (const sheet of document.styleSheets) {
    try {
      const rules = sheet.cssRules;
      for (const rule of rules) {
        if (
          "selectorText" in rule &&
          typeof rule.selectorText === "string" &&
          rule.selectorText.startsWith(".theme-")
        ) {
          const match = rule.selectorText.match(/\.theme-([a-zA-Z0-9_-]+)/);
          if (match && !seen.has(match[1])) {
            seen.add(match[1]);
            results.push(match[1]);
          }
        }
      }
    } catch (err) {
      // Cross-origin stylesheet or inaccessible
      continue;
    }
  }

  return results;
}
