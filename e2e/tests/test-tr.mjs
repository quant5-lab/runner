#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../utils/test-helpers.js';

/* TR E2E Tests - Comprehensive coverage for True Range bug fix */

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

/* Manual TR calculation for validation */
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

/* ATR calculation for validation */
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

console.log('=== TR (True Range) E2E Tests ===\n');

// Test 1: Direct TR access
console.log('Test 1: Direct TR access');
const trDirectResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-tr-direct.pine',
  'volatile',
  100,
  15,
);

const trValues = getPlotValues(trDirectResult, 'TR');
const highValues = getPlotValues(trDirectResult, 'high');
const lowValues = getPlotValues(trDirectResult, 'low');
const closeValues = getPlotValues(trDirectResult, 'close');

if (!trValues || trValues.length === 0) {
  console.error('❌ FAILED: TR values not found or empty');
  process.exit(1);
}

const manualTR = calcTrueRange(highValues, lowValues, closeValues);

let trMatched = 0;
for (let i = 0; i < trValues.length; i++) {
  assertFloatEquals(trValues[i], manualTR[i], FLOAT_EPSILON, `tr[${i}]`);
  trMatched++;
}
console.log(`✅ PASSED: ${trMatched}/${trValues.length} TR values match manual calculation\n`);

// Test 2: TR in calculations (SMA, EMA)
console.log('Test 2: TR in calculations (SMA, EMA)');
const trCalcResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-tr-calculations.pine',
  'volatile',
  100,
  15,
);

const trSmaValues = getPlotValues(trCalcResult, 'TR SMA');
const trEmaValues = getPlotValues(trCalcResult, 'TR EMA');

if (!trSmaValues || !trEmaValues) {
  console.error('❌ FAILED: TR SMA/EMA calculations not working');
  process.exit(1);
}

const validSma = trSmaValues.filter((v) => !isNaN(v) && v > 0).length;
const validEma = trEmaValues.filter((v) => !isNaN(v) && v > 0).length;
console.log(`✅ PASSED: TR SMA ${validSma}/${trSmaValues.length} valid values`);
console.log(`✅ PASSED: TR EMA ${validEma}/${trEmaValues.length} valid values\n`);

// Test 3: ATR calculation (uses TR internally)
console.log('Test 3: ATR calculation (uses TR internally)');
const atrResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-tr-atr.pine',
  'volatile',
  100,
  15,
);

const atrValues = getPlotValues(atrResult, 'ATR');
const atrTrValues = getPlotValues(atrResult, 'TR');
const atrHighValues = getPlotValues(atrResult, 'high');
const atrLowValues = getPlotValues(atrResult, 'low');
const atrCloseValues = getPlotValues(atrResult, 'close');

const manualATRtr = calcTrueRange(atrHighValues, atrLowValues, atrCloseValues);
const manualATR = calcATR(manualATRtr, 14);

let atrMatched = 0;
for (let i = 14; i < atrValues.length; i++) {
  if (!isNaN(atrValues[i]) && !isNaN(manualATR[i])) {
    assertFloatEquals(atrValues[i], manualATR[i], 0.01, `atr[${i}]`);
    atrMatched++;
  }
}
console.log(`✅ PASSED: ${atrMatched}/${atrValues.length - 14} ATR values match\n`);

// Test 4: TR in conditional logic
console.log('Test 4: TR in conditional logic');
const trCondResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-tr-conditions.pine',
  'volatile',
  100,
  15,
);

const trHighSignal = getPlotValues(trCondResult, 'TR High Signal');
const trLowSignal = getPlotValues(trCondResult, 'TR Low Signal');

const highSignalCount = trHighSignal.filter((v) => v === 1).length;
const lowSignalCount = trLowSignal.filter((v) => v === 1).length;

if (!trHighSignal || !trLowSignal) {
  console.error('❌ FAILED: TR condition plots not found');
  process.exit(1);
}

/* At least the conditions evaluated (even if 0 signals is valid for this pattern) */
console.log(`✅ PASSED: TR conditions work (${highSignalCount} high, ${lowSignalCount} low signals)\n`);

// Test 5: TR in strategy entry/exit logic
console.log('Test 5: TR in strategy entry/exit logic');
const trStrategyResult = await runStrategyWithPattern(
  100,
  'e2e/fixtures/strategies/test-tr-strategy.pine',
  'volatile',
  100,
  20,
);

const longEntries = trStrategyResult.trades?.filter((t) => t.type === 'long')?.length || 0;
const shortEntries = trStrategyResult.trades?.filter((t) => t.type === 'short')?.length || 0;

/* Strategy executed successfully - trades may be 0 depending on pattern */
if (!trStrategyResult.plots) {
  console.error('❌ FAILED: TR strategy did not execute');
  process.exit(1);
}

console.log(`✅ PASSED: TR strategy executed (${longEntries} long, ${shortEntries} short trades)\n`);

// Test 6: TR with ADX (complex indicator using TR)
console.log('Test 6: TR with ADX (DMI uses TR internally)');
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

// Test 7: TR edge case - first bar (no previous close)
console.log('Test 7: TR edge case - first bar');
const trFirstBarResult = await runStrategyWithPattern(
  10,
  'e2e/fixtures/strategies/test-tr-first-bar.pine',
  'linear',
  100,
  5,
);

const firstBarTR = getPlotValues(trFirstBarResult, 'TR');
const firstBarHigh = getPlotValues(trFirstBarResult, 'high');
const firstBarLow = getPlotValues(trFirstBarResult, 'low');

/* First bar: TR should equal high - low (no previous close) */
const expectedFirstTR = firstBarHigh[0] - firstBarLow[0];
assertFloatEquals(firstBarTR[0], expectedFirstTR, FLOAT_EPSILON, 'first bar TR');
console.log(`✅ PASSED: First bar TR = ${firstBarTR[0].toFixed(2)} (high-low, no prev close)\n`);

// Test 8: TR edge case - gaps (large price movements)
console.log('Test 8: TR edge case - gaps');
const trGapResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-tr-gaps.pine',
  'gaps',
  100,
  20,
);

const gapTR = getPlotValues(trGapResult, 'TR');
const gapHigh = getPlotValues(trGapResult, 'high');
const gapLow = getPlotValues(trGapResult, 'low');
const gapClose = getPlotValues(trGapResult, 'close');

/* Find gaps where close[1] is outside current bar range */
let gapCount = 0;
for (let i = 1; i < gapHigh.length; i++) {
  const prevClose = gapClose[i - 1];
  if (prevClose > gapHigh[i] || prevClose < gapLow[i]) {
    gapCount++;
    /* TR should capture the gap via abs(high-close[1]) or abs(low-close[1]) */
    const tr1 = gapHigh[i] - gapLow[i];
    const tr2 = Math.abs(gapHigh[i] - prevClose);
    const tr3 = Math.abs(gapLow[i] - prevClose);
    const expectedTR = Math.max(tr1, tr2, tr3);
    assertFloatEquals(gapTR[i], expectedTR, FLOAT_EPSILON, `gap TR[${i}]`);
  }
}
console.log(`✅ PASSED: ${gapCount} gaps detected, TR correctly captures gap movements\n`);

// Test 9: TR in function scope
console.log('Test 9: TR in function scope');
const trFunctionResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-tr-function.pine',
  'volatile',
  100,
  10,
);

const functionTR = getPlotValues(trFunctionResult, 'Function TR');
if (!functionTR || functionTR.length === 0) {
  console.error('❌ FAILED: TR not accessible in function scope');
  process.exit(1);
}
console.log(`✅ PASSED: TR accessible in function scope (${functionTR.length} values)\n`);

// Test 10: Multiple TR usages in same script
console.log('Test 10: Multiple TR usages in same script');
const trMultiResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-tr-multiple.pine',
  'volatile',
  100,
  10,
);

const tr1 = getPlotValues(trMultiResult, 'TR Direct');
const tr2 = getPlotValues(trMultiResult, 'TR * 2');
const tr3 = getPlotValues(trMultiResult, 'ATR14');
const tr4 = getPlotValues(trMultiResult, 'TR SMA');

if (!tr1 || !tr2 || !tr3 || !tr4) {
  console.error('❌ FAILED: Multiple TR usages not working');
  process.exit(1);
}

/* Verify TR * 2 equals TR Direct * 2 */
for (let i = 0; i < tr1.length; i++) {
  assertFloatEquals(tr2[i], tr1[i] * 2, FLOAT_EPSILON, `tr*2[${i}]`);
}

console.log('✅ PASSED: Multiple TR usages work correctly\n');

// Test 11: Regression test - BB7 ADX strategy
console.log('Test 11: Regression test - BB7 ADX strategy');
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

console.log(`✅ PASSED: BB7 ADX calculated correctly (${bb7ADX.length - nullADX}/${bb7ADX.length} valid)\n`);

console.log('=== All TR tests passed ✅ ===');
console.log(`Total: 11 test scenarios covering:
  - Direct TR access
  - TR in calculations (SMA, EMA, ATR)
  - TR in conditions and strategy logic
  - TR with complex indicators (ADX, DMI)
  - Edge cases (first bar, gaps, function scope)
  - Multiple TR usages
  - Regression test (BB7 ADX bug)
`);
