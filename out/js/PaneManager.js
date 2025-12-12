/* Multi-pane chart manager with time-scale synchronization (SRP) */
export class PaneManager {
  constructor(chartOptions) {
    this.chartOptions = chartOptions;
    this.mainPane = null;
    this.dynamicPanes = new Map();
  }

  createMainPane(container, config) {
    this.mainPane = {
      container,
      chart: LightweightCharts.createChart(container, {
        ...this.chartOptions,
        height: config.height,
        width: container.clientWidth,
      }),
    };
    return this.mainPane;
  }

  createDynamicPane(paneName, config) {
    const containerDiv = document.createElement('div');
    containerDiv.id = `${paneName}-chart`;
    containerDiv.style.position = 'relative';
    containerDiv.style.zIndex = '1';

    const chartContainerDiv = document.querySelector('.chart-container');
    chartContainerDiv.appendChild(containerDiv);

    const chart = LightweightCharts.createChart(containerDiv, {
      ...this.chartOptions,
      height: config.height,
      width: containerDiv.clientWidth,
    });

    this.dynamicPanes.set(paneName, { container: containerDiv, chart });
    return { container: containerDiv, chart };
  }

  getPane(paneName) {
    return paneName === 'main' ? this.mainPane : this.dynamicPanes.get(paneName);
  }

  getAllCharts() {
    const charts = [this.mainPane.chart];
    this.dynamicPanes.forEach(({ chart }) => charts.push(chart));
    return charts;
  }

  getAllContainers() {
    const containers = [this.mainPane.container];
    this.dynamicPanes.forEach(({ container }) => containers.push(container));
    return containers;
  }

  synchronizeTimeScales() {
    const charts = this.getAllCharts();
    let isUpdating = false;

    charts.forEach((sourceChart, sourceIndex) => {
      sourceChart.timeScale().subscribeVisibleLogicalRangeChange((logicalRange) => {
        if (isUpdating || !logicalRange) return;

        isUpdating = true;
        requestAnimationFrame(() => {
          charts.forEach((targetChart, targetIndex) => {
            if (sourceIndex !== targetIndex) {
              try {
                targetChart.timeScale().setVisibleLogicalRange(logicalRange);
              } catch (error) {
                console.warn('Failed to sync logical range:', error);
              }
            }
          });
          isUpdating = false;
        });
      });
    });
  }
}
