/* Lightweight Charts LineSeries Adapter (Adapter Pattern)
 * Adapts raw plot data with gaps to Lightweight Charts LineSeriesData format
 * Lightweight Charts v4.1.1 lacks native gap-breaking capability (Issue #699)
 */

/* Pure predicate: check if value is valid number */
const isValidValue = (value) => 
  value !== null && value !== undefined && !isNaN(value);

/* Pure predicate: check if point should be visible (has color) */
const hasColor = (item) => 
  item.options && item.options.color !== undefined && item.options.color !== null;

/* Pure function: convert milliseconds to seconds */
const msToSeconds = (ms) => Math.floor(ms / 1000);

/* Pure function: find first valid data point index */
const findFirstValidIndex = (data) => {
  for (let i = 0; i < data.length; i++) {
    if (isValidValue(data[i].value) && hasColor(data[i])) return i;
  }
  return -1;
};

/* Pure function: create invisible anchor point for alignment */
/* NaN prevents auto-scale inclusion (Lightweight Charts official pattern) */
const createAnchorPoint = (time) => ({
  time: msToSeconds(time),
  value: NaN,
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
  return next && (!isValidValue(next.value) || !hasColor(next));
};

/* Pure function: check if previous point was valid */
const prevIsValid = (data, index) => {
  if (index === 0) return false;
  const prev = data[index - 1];
  return prev && isValidValue(prev.value) && hasColor(prev);
};

/**
 * Adapt plot data to Lightweight Charts LineSeries format with gap handling
 * Strategy: Insert invisible anchors before first valid point, mark gap edges transparent,
 * and convert mid-series gaps to transparent points to break line continuity
 * Treats points without color (PineScript color=na) as gaps
 */
export function adaptLineSeriesData(plotData) {
  if (!Array.isArray(plotData)) return [];

  const firstValidIndex = findFirstValidIndex(plotData);
  if (firstValidIndex === -1) return [];

  return plotData.reduce((acc, item, i) => {
    const hasValidValue = isValidValue(item.value);
    const isVisible = hasColor(item);
    
    if (i < firstValidIndex) {
      acc.push(createAnchorPoint(item.time));
    } else if (hasValidValue && isVisible) {
      acc.push(createDataPoint(item.time, item.value, nextIsGap(plotData, i)));
    } else if (hasValidValue && !isVisible && prevIsValid(plotData, i)) {
      /* Point has value but no color (Pine color=na) - treat as gap */
      acc.push(createAnchorPoint(item.time));
    } else if (!hasValidValue && prevIsValid(plotData, i)) {
      /* Gap after valid point - add transparent NaN to break line */
      acc.push(createAnchorPoint(item.time));
    }
    return acc;
  }, []);
}
