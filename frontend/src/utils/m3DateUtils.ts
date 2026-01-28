/**
 * M3 Date Utilities
 * Helper functions for working with M3 date formats (YYYYMMDD integers)
 */

/**
 * Calculate the absolute difference in days between two M3 dates
 * @param date1 M3 date in YYYYMMDD format (number or string)
 * @param date2 M3 date in YYYYMMDD format (number or string)
 * @returns Absolute number of days between the two dates
 */
export function dateDiffDays(date1: number | string, date2: number | string): number {
  const str1 = date1.toString();
  const str2 = date2.toString();

  const d1 = new Date(
    parseInt(str1.substring(0, 4)),
    parseInt(str1.substring(4, 6)) - 1,
    parseInt(str1.substring(6, 8))
  );
  const d2 = new Date(
    parseInt(str2.substring(0, 4)),
    parseInt(str2.substring(4, 6)) - 1,
    parseInt(str2.substring(6, 8))
  );

  return Math.abs(Math.floor((d2.getTime() - d1.getTime()) / (1000 * 60 * 60 * 24)));
}

/**
 * Get Tailwind CSS classes for variance badge color based on day count
 * @param days Number of days variance
 * @returns Tailwind CSS classes for background, text, and ring colors
 */
export function getVarianceBadgeColor(days: number): string {
  if (days > 7) return 'bg-red-100 text-red-800 ring-red-300';
  if (days >= 3) return 'bg-orange-100 text-orange-800 ring-orange-300';
  return 'bg-yellow-100 text-yellow-800 ring-yellow-300';
}
