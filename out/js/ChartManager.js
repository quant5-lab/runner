export class ChartManager {
  static createChart(container, config, chartOptions) {
    return LightweightCharts.createChart(container, {
      ...chartOptions,
      height: config.height,
      width: container.clientWidth,
    });
  }

  static addCandlestickSeries(chart, config) {
    return chart.addCandlestickSeries(config);
  }

  static addLineSeries(chart, config) {
    return chart.addLineSeries(config);
  }

  static addHistogramSeries(chart, config) {
    return chart.addHistogramSeries(config);
  }

  static fitContent(charts) {
    charts.forEach((chart) => chart.timeScale().fitContent());
  }

  static handleResize(charts, containers) {
    charts.forEach((chart, i) => {
      chart.resize(containers[i].clientWidth, containers[i].clientHeight);
    });
  }
}
