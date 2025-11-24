#!/usr/bin/env node
import { createContainer } from '../src/container.js';
import { createProviderChain, DEFAULTS } from '../src/config.js';
import { writeFile } from 'fs/promises';
import { dirname } from 'path';
import { mkdir } from 'fs/promises';

async function fetchGDYNData() {
  const container = createContainer(createProviderChain, DEFAULTS);
  const logger = container.resolve('logger');
  const providerManager = container.resolve('providerManager');

  logger.info('Fetching GDYN 1h data via provider chain...');

  const symbol = 'GDYN';
  const timeframe = '1h';
  const bars = 1000;

  try {
    const data = await providerManager.getMarketData(symbol, timeframe, bars);

    if (!data || data.length === 0) {
      logger.error('No data received from provider chain');
      process.exit(1);
    }

    logger.info(`Received ${data.length} bars`);

    /* Convert to golang-port format: {time, open, high, low, close, volume} */
    const golangFormat = data.map(bar => ({
      time: Math.floor(bar.openTime / 1000),
      open: bar.open,
      high: bar.high,
      low: bar.low,
      close: bar.close,
      volume: bar.volume
    }));

    const outputPath = './golang-port/testdata/gdyn-1h.json';

    /* Ensure directory exists */
    await mkdir(dirname(outputPath), { recursive: true });

    await writeFile(outputPath, JSON.stringify(golangFormat, null, 2));

    logger.info(`✓ Data saved to ${outputPath}`);
    logger.info(`  Bars: ${golangFormat.length}`);
    logger.info(`  First bar time: ${new Date(golangFormat[0].time * 1000).toISOString()}`);
    logger.info(`  Last bar time: ${new Date(golangFormat[golangFormat.length - 1].time * 1000).toISOString()}`);

    const stats = container.resolve('apiStatsCollector');
    stats.logSummary(logger);
  } catch (error) {
    logger.error('Error fetching data:', error);
    process.exit(1);
  }
}

fetchGDYNData();
