#!/usr/bin/env node
/* Generate OHLCV test data from Node.js providers for golang-port testing */

import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
import { MoexProvider } from '../src/providers/MoexProvider.js';
import { BinanceProvider } from '../src/providers/BinanceProvider.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const OUTPUT_DIR = path.join(__dirname, '../golang-port/testdata/ohlcv');

/* Fetch and save data */
async function fetchAndSave(provider, symbol, timeframe, limit, filename) {
  console.log(`Fetching ${symbol} ${timeframe} (${limit} bars)...`);
  const data = await provider.getMarketData(symbol, timeframe, limit);
  
  if (!data || data.length === 0) {
    console.warn(`⚠ Warning: ${symbol} ${timeframe} returned 0 bars`);
  }
  
  const outputPath = path.join(OUTPUT_DIR, filename);
  fs.writeFileSync(outputPath, JSON.stringify(data, null, 2));
  console.log(`✓ Saved ${data.length} bars to ${filename}`);
  return data.length;
}

async function main() {
  /* Create dummy logger and stats collector */
  const logger = {
    debug: (...args) => {},
    info: (...args) => console.log(...args),
    error: (...args) => console.error(...args),
  };
  const statsCollector = {
    recordCacheHit: () => {},
    recordCacheMiss: () => {},
    recordApiCall: () => {},
    recordRequest: () => {},
  };
  
  const moex = new MoexProvider(logger, statsCollector);
  const binance = new BinanceProvider(logger, statsCollector);
  
  console.log('=== Generating Test Data ===\n');
  
  try {
    /* MOEX: GAZP (Gazprom - large liquid stock) */
    await fetchAndSave(moex, 'GAZP', '1h', 500, 'GAZP_1h.json');
    await fetchAndSave(moex, 'GAZP', '1D', 1020, 'GAZP_1D.json');
    
    /* MOEX: SBER */
    await fetchAndSave(moex, 'SBER', '1h', 500, 'SBER_1h.json');
    await fetchAndSave(moex, 'SBER', '1D', 1020, 'SBER_1D.json');
    
    /* Binance: BTCUSDT (already exists, regenerate for consistency) */
    await fetchAndSave(binance, 'BTCUSDT', '1h', 500, 'BTCUSDT_1h.json');
    await fetchAndSave(binance, 'BTCUSDT', '1D', 1020, 'BTCUSDT_1D.json');
    
    console.log('\n✓ Test data generation complete');
  } catch (error) {
    console.error(`✗ Error: ${error.message}`);
    process.exit(1);
  }
}

main();
