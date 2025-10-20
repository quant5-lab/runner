#!/usr/bin/env node
/**
 * E2E Test: plot() parameters with DETERMINISTIC data validation
 *
 * Tests that all plot() parameters are passed through correctly:
 * 1. Basic params: color, linewidth, style
 * 2. Transparency: transp
 * 3. Histogram params: histbase, offset
 */
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test: plot() Parameters with Deterministic Data');
console.log('═══════════════════════════════════════════════════════════\n');

// Create container with MockProvider
const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

// Read and transpile strategy
const pineCode = await readFile('e2e/fixtures/strategies/test-plot-params.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);

// Run strategy with deterministic data (30 bars)
const result = await runner.runPineScriptStrategy(
  'TEST',
  'D',
  30,
  jsCode,
  'test-plot-params.pine',
);

console.log('=== DETERMINISTIC TEST RESULTS ===\n');

// Test 1: Verify SMA20 plot has basic params
console.log('TEST 1: SMA20 plot basic parameters');
const sma20Plot = result.plots?.['SMA20'];
if (!sma20Plot) {
  console.error('❌ FAILED: SMA20 plot not found');
  process.exit(1);
}

const sma20Options = sma20Plot.data?.[0]?.options || {};
console.log('  SMA20 options:', JSON.stringify(sma20Options, null, 2));

if (sma20Options.color !== 'blue') {
  console.error(`❌ FAILED: Expected color='blue', got '${sma20Options.color}'`);
  process.exit(1);
}
if (sma20Options.linewidth !== 2) {
  console.error(`❌ FAILED: Expected linewidth=2, got ${sma20Options.linewidth}`);
  process.exit(1);
}
console.log('✅ PASSED: SMA20 has correct color, linewidth (style is identifier, checked separately)\n');

// Test 2: Verify Close plot has transp parameter
console.log('TEST 2: Close plot transparency parameter');
const closePlot = result.plots?.['Close'];
if (!closePlot) {
  console.error('❌ FAILED: Close plot not found');
  process.exit(1);
}

const closeOptions = closePlot.data?.[0]?.options || {};
console.log('  Close options:', JSON.stringify(closeOptions, null, 2));

if (closeOptions.color !== 'red') {
  console.error(`❌ FAILED: Expected color='red', got '${closeOptions.color}'`);
  process.exit(1);
}
if (closeOptions.linewidth !== 1) {
  console.error(`❌ FAILED: Expected linewidth=1, got ${closeOptions.linewidth}`);
  process.exit(1);
}
if (closeOptions.transp !== 50) {
  console.error(`❌ FAILED: Expected transp=50, got ${closeOptions.transp}`);
  process.exit(1);
}
console.log('✅ PASSED: Close plot has correct transp parameter\n');

// Test 3: Verify Volume plot has histbase and offset
console.log('TEST 3: Volume plot histogram parameters');
const volumePlot = result.plots?.['Volume'];
if (!volumePlot) {
  console.error('❌ FAILED: Volume plot not found');
  process.exit(1);
}

const volumeOptions = volumePlot.data?.[0]?.options || {};
console.log('  Volume options:', JSON.stringify(volumeOptions, null, 2));

if (volumeOptions.color !== 'green') {
  console.error(`❌ FAILED: Expected color='green', got '${volumeOptions.color}'`);
  process.exit(1);
}
if (volumeOptions.histbase !== 0) {
  console.error(`❌ FAILED: Expected histbase=0, got ${volumeOptions.histbase}`);
  process.exit(1);
}
if (volumeOptions.offset !== 1) {
  console.error(`❌ FAILED: Expected offset=1, got ${volumeOptions.offset}`);
  process.exit(1);
}
console.log('✅ PASSED: Volume plot has correct histbase and offset parameters (style is identifier)\n');

console.log('═══════════════════════════════════════════════════════════');
console.log('✅ ALL TESTS PASSED: plot() parameters correctly passed through');
console.log('═══════════════════════════════════════════════════════════');
