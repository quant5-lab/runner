/* Lightweight Charts LineSeries Adapter (Adapter Pattern)
 * Adapts raw plot data with gaps to Lightweight Charts LineSeriesData format
 *
 * Lightweight Charts library (v4.1.1) lacks native gap-breaking capability:
 * - No connectGaps/connectWhitespace option exists
 * - Library interpolates across missing data points by default
 * - GitHub Issue #699: Feature request open since Feb 2021
 *
 * Adapter strategy:
 * 1. Filter out null/undefined/NaN values (removes gaps from dataset)
 * 2. Mark transition edges with transparent color (prevents interpolation)
 *
 * When a valid data point is followed by null:
 * - Set color='transparent' on the last valid point
 * - Prevents line drawing to the next segment after the gap
 * - Creates visual appearance of line breaking at position close
 */

/**
 * Adapt plot data to Lightweight Charts LineSeries format with gap handling
 * @param {Array<{time: number, value: number|null}>} plotData - Raw plot data with timestamps and values
 * @returns {Array<{time: number, value: number, color?: string}>} - Adapted data with transparent edges
 */
export function adaptLineSeriesData(plotData) {
  if (!Array.isArray(plotData)) {
    return [];
  }

  const filtered = [];

  for (let i = 0; i < plotData.length; i++) {
    const item = plotData[i];
    const hasValue = item.value !== null && item.value !== undefined && !isNaN(item.value);

    if (hasValue) {
      const point = {
        time: Math.floor(item.time / 1000),
        value: item.value,
      };

      /* Check if next point is null - mark edge as transparent */
      const nextItem = plotData[i + 1];
      if (nextItem) {
        const nextHasValue =
          nextItem.value !== null && nextItem.value !== undefined && !isNaN(nextItem.value);

        if (!nextHasValue) {
          point.color = 'transparent';
        }
      }

      filtered.push(point);
    }
  }

  return filtered;
}
