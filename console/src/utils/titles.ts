// console/src/utils/titles.ts

export const newspaperTitles = [
  "The VTR Times",
  "The VTR Dispatch",
  "The VTR Bulletin",
  "The VTR Feedline",
  "The VTR Signal",
  "Feeder Solutions Chronicle",
  "Feeder Solutions Digest",
  "The Feeder Report",
  "The Feeder Insight",
  "VTR View",
];
export function getRandomTitle(): string {
  return newspaperTitles[Math.floor(Math.random() * newspaperTitles.length)];
}

export function getTitleByIndex(index: number): string {
  return newspaperTitles[index % newspaperTitles.length];
}

export function getAllTitles(): string[] {
  return [...newspaperTitles];
}
