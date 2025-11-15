#!/usr/bin/env node
import { createContainer } from '../../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../../mocks/MockProvider.js';

console.log('=== TA Function Test: time() ===\n');

const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

const pineCode = await readFile('e2e/fixtures/strategies/test-time.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);
const result = await runner.runPineScriptStrategy('TEST', '1h', 30, jsCode, 'test-time.pine');

function getPlotValues(result, plotTitle) {
  const plot = result.plots?.[plotTitle];
  if (!plot || !plot.data) return null;
  return plot.data.map((d) => d.value);
}

const timeDailyValues = getPlotValues(result, 'time_daily');
const timeWeeklyValues = getPlotValues(result, 'time_weekly');

const validDailyTimes = timeDailyValues.filter((v) => v !== null && !isNaN(v) && v > 0).length;
const validWeeklyTimes = timeWeeklyValues.filter((v) => v !== null && !isNaN(v) && v > 0).length;

console.log(
  `✅ time: Daily ${validDailyTimes}/${timeDailyValues.length}, Weekly ${validWeeklyTimes}/${timeWeeklyValues.length} valid timestamps`
);

process.exit(0);
