import { writeFileSync, mkdirSync } from 'fs';
import { join } from 'path';

class JsonFileWriter {
  constructor(logger) {
    this.logger = logger;
  }

  ensureOutDirectory() {
    try {
      mkdirSync('out', { recursive: true });
    } catch (error) {
      this.logger.debug(`Failed to create output directory: ${error.message}`);
    }
  }

  exportChartData(candlestickData, plots, strategy = null) {
    this.ensureOutDirectory();
    const chartData = {
      candlestick: candlestickData,
      plots,
      timestamp: new Date().toISOString(),
    };
    
    if (strategy && (strategy.trades?.length > 0 || strategy.openTrades?.length > 0)) {
      chartData.strategy = strategy;
    }

    writeFileSync(join('out', 'chart-data.json'), JSON.stringify(chartData, null, 2));
  }

  exportConfiguration(config) {
    this.ensureOutDirectory();
    writeFileSync(join('out', 'chart-config.json'), JSON.stringify(config, null, 2));
  }
}

export { JsonFileWriter };
