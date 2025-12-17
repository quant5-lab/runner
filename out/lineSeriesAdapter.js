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

/* Normalize timestamp to seconds (handles both seconds and milliseconds) */
const toSeconds = (time) => time > 10000000000 ? Math.floor(time / 1000) : time;

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
  time: toSeconds(time),
  value: NaN,
  color: 'transparent',
});

/* Pure function: create chart data point */
/* Note: Previously marked gap edges as transparent, but this caused rendering
 * issues with short segments (e.g., 5 points) where the last point would become
 * invisible. Gap handling is now done solely through anchor points (NaN values). */
const createDataPoint = (time, value) => ({
  time: toSeconds(time),
  value
});

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
function adaptLineSeriesData(plotData) {
  if (!Array.isArray(plotData)) return [];

  const firstValidIndex = findFirstValidIndex(plotData);
  if (firstValidIndex === -1) return [];

  return plotData.reduce((acc, item, i) => {
    const hasValidValue = isValidValue(item.value);
    const isVisible = hasColor(item);
    
    if (i < firstValidIndex) {
      acc.push(createAnchorPoint(item.time));
    } else if (hasValidValue && isVisible) {
      acc.push(createDataPoint(item.time, item.value));
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

/* Export to window for compatibility */
window.adaptLineSeriesData = adaptLineSeriesData;
