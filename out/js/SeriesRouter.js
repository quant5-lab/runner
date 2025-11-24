/* Series routing to correct panes (SRP) */
export class SeriesRouter {
  constructor(paneManager, seriesMap) {
    this.paneManager = paneManager;
    this.seriesMap = seriesMap;
  }

  routeSeries(seriesKey, seriesConfig, chartManager) {
    const paneName = seriesConfig.chart || 'indicator';
    const pane = this.paneManager.getPane(paneName);

    if (!pane) {
      console.warn(`Pane '${paneName}' not found for series '${seriesKey}'`);
      return null;
    }

    const seriesType = seriesConfig.style || 'line';
    let series;

    if (seriesType === 'histogram') {
      series = chartManager.addHistogramSeries(pane.chart, seriesConfig);
    } else {
      series = chartManager.addLineSeries(pane.chart, seriesConfig);
    }

    this.seriesMap[seriesKey] = series;
    return series;
  }

  rerouteSeries(seriesKey, newPaneName, seriesConfig, chartManager) {
    const oldSeries = this.seriesMap[seriesKey];
    if (!oldSeries) return null;

    const oldPaneName = seriesConfig.chart;
    const oldPane = this.paneManager.getPane(oldPaneName);

    if (oldPane && oldPane.chart) {
      oldPane.chart.removeSeries(oldSeries);
    }

    seriesConfig.chart = newPaneName;
    return this.routeSeries(seriesKey, seriesConfig, chartManager);
  }
}
