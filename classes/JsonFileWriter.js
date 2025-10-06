import { writeFileSync, mkdirSync } from 'fs';
import { join } from 'path';

class JsonFileWriter {
  ensureOutDirectory() {
    try {
      mkdirSync('out', { recursive: true });
    } catch (error) {
      /* Directory already exists */
    }
  }

  exportChartData(candlestickData, plots) {
    this.ensureOutDirectory();
    const chartData = {
      candlestick: candlestickData,
      plots,
      timestamp: new Date().toISOString(),
    };

    writeFileSync(join('out', 'chart-data.json'), JSON.stringify(chartData, null, 2));
  }

  exportConfiguration(config) {
    this.ensureOutDirectory();
    writeFileSync(join('out', 'chart-config.json'), JSON.stringify(config, null, 2));
  }
}

export { JsonFileWriter };
