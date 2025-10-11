/* Pine Script plot() adapter - bridges PyneScript transpiler output to PineTS API
 * 
 * PyneScript transpiles: plot(series, color=X, title=Y)
 * Into: plot(series, {color: X, title: Y})
 * 
 * PineTS expects: plot(series, title, options)
 * 
 * This adapter extracts title from options object and calls PineTS correctly.
 */

export const plotAdapterSource = `function plot(series, titleOrOptions, maybeOptions) {
  if (typeof titleOrOptions === 'string') {
    return corePlot(series, titleOrOptions, maybeOptions || {});
  }
  return corePlot(
    series,
    ((titleOrOptions && titleOrOptions[0]) || titleOrOptions || {}).title,
    {
      color: ((titleOrOptions && titleOrOptions[0]) || titleOrOptions || {}).color,
      style: ((titleOrOptions && titleOrOptions[0]) || titleOrOptions || {}).style,
      linewidth: ((titleOrOptions && titleOrOptions[0]) || titleOrOptions || {}).linewidth
    }
  );
}`;

/**
 * Create plot adapter function for testing
 * @param {Function} corePlot - PineTS core plot function
 * @returns {Function} - Adapted plot function
 */
export function createPlotAdapter(corePlot) {
  const fn = new Function('corePlot', `${plotAdapterSource}; return plot;`);
  return fn(corePlot);
}
