/* Lightweight Charts LineSeries Adapter (Adapter Pattern)
 * Adapts raw plot data with gaps to Lightweight Charts LineSeriesData format
 * Lightweight Charts v4.1.1 lacks native gap-breaking capability (Issue #699)
 */

/* Pure predicate: check if value is valid number */
const isValidValue = (value) => 
  value !== null && value !== undefined && !isNaN(value);

/* Pure function: convert milliseconds to seconds */
const msToSeconds = (ms) => Math.floor(ms / 1000);

/* Pure function: find first valid data point index */
const findFirstValidIndex = (data) => {
  for (let i = 0; i < data.length; i++) {
    if (isValidValue(data[i].value)) return i;
  }
  return -1;
};

/* Pure function: create invisible anchor point for alignment */
const createAnchorPoint = (time) => ({
  time: msToSeconds(time),
  value: 0,
  color: 'transparent',
});

/* Pure function: create chart data point with optional gap edge marking */
const createDataPoint = (time, value, isGapEdge) => {
  const point = { time: msToSeconds(time), value };
  if (isGapEdge) point.color = 'transparent';
  return point;
};

/* Pure function: check if next point starts a gap */
const nextIsGap = (data, index) => {
  const next = data[index + 1];
  return next && !isValidValue(next.value);
};

/**
 * Adapt plot data to Lightweight Charts LineSeries format with gap handling
 * Strategy: Insert invisible anchors before first valid point, mark gap edges transparent
 */
export function adaptLineSeriesData(plotData) {
  if (!Array.isArray(plotData)) return [];

  const firstValidIndex = findFirstValidIndex(plotData);
  if (firstValidIndex === -1) return [];

  return plotData.reduce((acc, item, i) => {
    if (i < firstValidIndex) {
      acc.push(createAnchorPoint(item.time));
    } else if (isValidValue(item.value)) {
      acc.push(createDataPoint(item.time, item.value, nextIsGap(plotData, i)));
    }
    return acc;
  }, []);
}
