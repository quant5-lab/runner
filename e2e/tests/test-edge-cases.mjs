#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../utils/test-helpers.js';

/* Edge Cases E2E Tests - First bar, gaps, discontinuities */

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

console.log('=== Edge Cases E2E Tests ===\n');

// Test 1: First bar behavior - no history available
console.log('Test 1: First bar behavior - no history available');
const firstBarResult = await runStrategyWithPattern(
  10,
  'e2e/fixtures/strategies/test-edge-first-bar.pine',
  'linear',
  100,
  5,
);

const firstClose = getPlotValues(firstBarResult, 'close');
const firstHl2 = getPlotValues(firstBarResult, 'hl2');
const firstTr = getPlotValues(firstBarResult, 'tr');
const firstHigh = getPlotValues(firstBarResult, 'high');
const firstLow = getPlotValues(firstBarResult, 'low');

/* First bar: close[1] doesn't exist, tr should be high - low */
const expectedFirstTR = firstHigh[0] - firstLow[0];
assertFloatEquals(firstTr[0], expectedFirstTR, FLOAT_EPSILON, 'first bar TR');

/* First bar: hl2 should work normally */
const expectedFirstHl2 = (firstHigh[0] + firstLow[0]) / 2;
assertFloatEquals(firstHl2[0], expectedFirstHl2, FLOAT_EPSILON, 'first bar hl2');

console.log('✅ PASSED: First bar edge case handled correctly\n');

// Test 2: Gap detection - discontinuities in price
console.log('Test 2: Gap detection - discontinuities in price');
const gapResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-edge-gaps.pine',
  'gaps',
  100,
  20,
);

const gapClose = getPlotValues(gapResult, 'close');
const gapHigh = getPlotValues(gapResult, 'high');
const gapLow = getPlotValues(gapResult, 'low');
const gapTr = getPlotValues(gapResult, 'tr');
const gapDetected = getPlotValues(gapResult, 'gap_detected');

/* Count gaps where close[1] is outside current bar range */
let gapCount = 0;
for (let i = 1; i < gapHigh.length; i++) {
  const prevClose = gapClose[i - 1];
  if (prevClose > gapHigh[i] || prevClose < gapLow[i]) {
    gapCount++;
    /* TR should capture gap via abs(high-close[1]) or abs(low-close[1]) */
    const tr1 = gapHigh[i] - gapLow[i];
    const tr2 = Math.abs(gapHigh[i] - prevClose);
    const tr3 = Math.abs(gapLow[i] - prevClose);
    const expectedTR = Math.max(tr1, tr2, tr3);
    assertFloatEquals(gapTr[i], expectedTR, FLOAT_EPSILON, `gap TR[${i}]`);
  }
}

const detectedCount = gapDetected.filter((v) => v === 1).length;
console.log(`✅ PASSED: ${gapCount} gaps detected, TR captures discontinuities correctly\n`);

// Test 3: Zero/negative values handling
console.log('Test 3: Zero/negative values handling');
const edgeValResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-edge-values.pine',
  'edge',
  100,
  10,
);

const edgeClose = getPlotValues(edgeValResult, 'close');
const edgeVolume = getPlotValues(edgeValResult, 'volume');
const edgeTr = getPlotValues(edgeValResult, 'tr');

/* All variables should have valid values (no undefined, null) */
let validCount = 0;
for (let i = 0; i < edgeClose.length; i++) {
  if (edgeClose[i] !== null && edgeClose[i] !== undefined &&
      edgeVolume[i] !== null && edgeVolume[i] !== undefined &&
      edgeTr[i] !== null && edgeTr[i] !== undefined) {
    validCount++;
  }
}

if (validCount !== edgeClose.length) {
  console.error(`❌ FAILED: Invalid values found (${validCount}/${edgeClose.length} valid)`);
  process.exit(1);
}

console.log(`✅ PASSED: Edge values handled correctly (${validCount}/${edgeClose.length} valid)\n`);

console.log('=== All edge case tests passed ✅ ===');
console.log(`Total: 3 test scenarios covering:
  - First bar behavior (no historical data)
  - Gap detection (price discontinuities)
  - Edge value handling (zero/negative values)
`);
