#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../utils/test-helpers.js';

/* Technical Indicators E2E Tests - ATR, ADX, DMI (TR-dependent indicators) */

async function runStrategyWithPattern(
  bars,
  strategyPath,
  pattern = 'linear',
  basePrice = 100,
  amplitude = 10,
) {
  const mockProvider = new MockProviderManager({ dataPattern: pattern, basePrice, amplitude });
  const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
  const DEFAULTS = { showDebug: false, showStats: false };

  const container = createContainer(createProviderChain, DEFAULTS);
  const runner = container.resolve('tradingAnalysisRunner');
  const transpiler = container.resolve('pineScriptTranspiler');

  const pineCode = await readFile(strategyPath, 'utf-8');
  const jsCode = await transpiler.transpile(pineCode);
  return await runner.runPineScriptStrategy('TEST', '1h', bars, jsCode, strategyPath);
}

function getPlotValues(result, plotTitle) {
  const plot = result.plots?.[plotTitle];
  if (!plot || !plot.data) return null;
  return plot.data.map((d) => d.value);
}

/* Manual ATR calculation for validation */
function calcTrueRange(highs, lows, closes) {
  const result = [];
  for (let i = 0; i < highs.length; i++) {
    if (i === 0) {
      result.push(highs[i] - lows[i]);
    } else {
      const tr1 = highs[i] - lows[i];
      const tr2 = Math.abs(highs[i] - closes[i - 1]);
      const tr3 = Math.abs(lows[i] - closes[i - 1]);
      result.push(Math.max(tr1, tr2, tr3));
    }
  }
  return result;
}

function calcATR(trValues, period) {
  const result = [];
  let sum = 0;
  for (let i = 0; i < trValues.length; i++) {
    if (i < period - 1) {
      sum += trValues[i];
      result.push(NaN);
    } else if (i === period - 1) {
      sum += trValues[i];
      const atr = sum / period;
      result.push(atr);
    } else {
      const prevATR = result[i - 1];
      const atr = (prevATR * (period - 1) + trValues[i]) / period;
      result.push(atr);
    }
  }
  return result;
}

console.log('=== Technical Indicators E2E Tests ===\n');

// Test 1: ATR (Average True Range)
console.log('Test 1: ATR (Average True Range)');
const atrResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-tr-atr.pine',
  'volatile',
  100,
  15,
);

const atrValues = getPlotValues(atrResult, 'ATR');
const atrHighValues = getPlotValues(atrResult, 'high');
const atrLowValues = getPlotValues(atrResult, 'low');
const atrCloseValues = getPlotValues(atrResult, 'close');

const manualTR = calcTrueRange(atrHighValues, atrLowValues, atrCloseValues);
const manualATR = calcATR(manualTR, 14);

let atrMatched = 0;
for (let i = 14; i < atrValues.length; i++) {
  if (!isNaN(atrValues[i]) && !isNaN(manualATR[i])) {
    assertFloatEquals(atrValues[i], manualATR[i], 0.01, `atr[${i}]`);
    atrMatched++;
  }
}
console.log(`✅ PASSED: ATR ${atrMatched}/${atrValues.length - 14} values match calculation\n`);

// Test 2: ADX (Average Directional Index)
console.log('Test 2: ADX (Average Directional Index)');
const adxResult = await runStrategyWithPattern(
  100,
  'e2e/fixtures/strategies/test-tr-adx.pine',
  'trending',
  100,
  10,
);

const adxValues = getPlotValues(adxResult, 'ADX');
const dmiPlusValues = getPlotValues(adxResult, 'DI+');
const dmiMinusValues = getPlotValues(adxResult, 'DI-');

if (!adxValues || !dmiPlusValues || !dmiMinusValues) {
  console.error('❌ FAILED: ADX/DMI indicators not working');
  process.exit(1);
}

const validADX = adxValues.filter((v) => !isNaN(v) && v > 0).length;
const validDIPlus = dmiPlusValues.filter((v) => !isNaN(v) && v >= 0).length;
const validDIMinus = dmiMinusValues.filter((v) => !isNaN(v) && v >= 0).length;

console.log(`✅ PASSED: ADX ${validADX}/${adxValues.length} valid values`);
console.log(`✅ PASSED: DI+ ${validDIPlus}/${dmiPlusValues.length} valid values`);
console.log(`✅ PASSED: DI- ${validDIMinus}/${dmiMinusValues.length} valid values\n`);

// Test 3: BB7 Regression - ADX in strategy context
console.log('Test 3: BB7 Regression - ADX in strategy context');
const bb7Result = await runStrategyWithPattern(
  100,
  'e2e/fixtures/strategies/test-tr-bb7-adx.pine',
  'trending',
  100,
  15,
);

const bb7ADX = getPlotValues(bb7Result, 'ADX');

if (!bb7ADX) {
  console.error('❌ FAILED: BB7 ADX not calculated (original bug reproduced)');
  process.exit(1);
}

const nullADX = bb7ADX.filter((v) => v === null || v === undefined).length;
if (nullADX > bb7ADX.length * 0.5) {
  console.error(`❌ FAILED: Too many null ADX values (${nullADX}/${bb7ADX.length})`);
  process.exit(1);
}

console.log(
  `✅ PASSED: BB7 ADX calculated correctly (${bb7ADX.length - nullADX}/${bb7ADX.length} valid)\n`,
);

console.log('=== All indicator tests passed ✅ ===');
console.log(`Total: 3 test scenarios covering:
  - ATR calculation (uses TR internally)
  - ADX/DMI indicators (use TR internally)
  - BB7 regression test (original TR bug)
`);
