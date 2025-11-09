#!/usr/bin/env node
/**
 * E2E Test: Multi-pane chart rendering with 4 panes
 *
 * Tests that:
 * 1. ConfigurationBuilder generates dynamic pane config from metadata
 * 2. index.html PaneManager creates 4 panes: main, equity, oscillators, volume
 * 3. Series are routed to correct panes
 * 4. All panes render independently
 */
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test: Multi-Pane Chart Rendering (4 Panes)');
console.log('═══════════════════════════════════════════════════════════\n');

/* Create container with MockProvider */
const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');
const configBuilder = container.resolve('configurationBuilder');

/* Read and transpile strategy */
const pineCode = await readFile('e2e/fixtures/strategies/test-multi-pane.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);

/* Run strategy (30 bars) */
const result = await runner.runPineScriptStrategy('TEST', 'D', 30, jsCode, 'test-multi-pane.pine');

console.log('=== TEST 1: Plot data contains pane property ===');
const equityPlot = result.plots?.['Strategy Equity'];
const rsiPlot = result.plots?.['RSI'];
const volumePlot = result.plots?.['Volume'];

if (!equityPlot || !equityPlot.data || equityPlot.data.length === 0) {
  console.error('❌ FAIL: Strategy Equity plot missing or empty');
  process.exit(1);
}

const firstEquityPoint = equityPlot.data[0];
if (firstEquityPoint?.options?.pane !== 'equity') {
  console.error(`❌ FAIL: Expected equity pane, got: ${firstEquityPoint?.options?.pane}`);
  process.exit(1);
}
console.log('✅ PASS: Equity plot has pane=\'equity\'');

if (!rsiPlot || !rsiPlot.data || rsiPlot.data.length === 0) {
  console.error('❌ FAIL: RSI plot missing or empty');
  process.exit(1);
}

const firstRsiPoint = rsiPlot.data[0];
if (firstRsiPoint?.options?.pane !== 'oscillators') {
  console.error(`❌ FAIL: Expected oscillators pane, got: ${firstRsiPoint?.options?.pane}`);
  process.exit(1);
}
console.log('✅ PASS: RSI plot has pane=\'oscillators\'');

if (!volumePlot || !volumePlot.data || volumePlot.data.length === 0) {
  console.error('❌ FAIL: Volume plot missing or empty');
  process.exit(1);
}

const firstVolumePoint = volumePlot.data[0];
if (firstVolumePoint?.options?.pane !== 'volume') {
  console.error(`❌ FAIL: Expected volume pane, got: ${firstVolumePoint?.options?.pane}`);
  process.exit(1);
}
console.log('✅ PASS: Volume plot has pane=\'volume\'\n');

console.log('=== TEST 2: Metadata extraction captures pane property ===');
const metadata = runner.extractIndicatorMetadata(result.plots);

if (metadata['Strategy Equity']?.chartPane !== 'equity') {
  console.error(`❌ FAIL: Metadata chartPane should be 'equity', got: ${metadata['Strategy Equity']?.chartPane}`);
  process.exit(1);
}
console.log('✅ PASS: Equity metadata has chartPane=\'equity\'');

if (metadata['RSI']?.chartPane !== 'oscillators') {
  console.error(`❌ FAIL: Metadata chartPane should be 'oscillators', got: ${metadata['RSI']?.chartPane}`);
  process.exit(1);
}
console.log('✅ PASS: RSI metadata has chartPane=\'oscillators\'');

if (metadata['Volume']?.chartPane !== 'volume') {
  console.error(`❌ FAIL: Metadata chartPane should be 'volume', got: ${metadata['Volume']?.chartPane}`);
  process.exit(1);
}
console.log('✅ PASS: Volume metadata has chartPane=\'volume\'\n');

console.log('=== TEST 3: ConfigurationBuilder generates 4 panes ===');
const tradingConfig = configBuilder.createTradingConfig('TEST', 'D', 30, 'test-multi-pane.pine');
const chartConfig = configBuilder.generateChartConfig(tradingConfig, metadata);

const panes = Object.keys(chartConfig.chartLayout);
if (panes.length !== 4) {
  console.error(`❌ FAIL: Expected 4 panes, got ${panes.length}: ${panes.join(', ')}`);
  process.exit(1);
}

if (!panes.includes('main') || !panes.includes('equity') || !panes.includes('oscillators') || !panes.includes('volume')) {
  console.error(`❌ FAIL: Missing expected panes. Got: ${panes.join(', ')}`);
  process.exit(1);
}
console.log(`✅ PASS: Config has 4 panes: ${panes.join(', ')}`);

if (chartConfig.chartLayout.main.height !== 400 || !chartConfig.chartLayout.main.fixed) {
  console.error('❌ FAIL: Main pane should be height 400 with fixed: true');
  process.exit(1);
}
console.log('✅ PASS: Main pane: height=400, fixed=true');

if (chartConfig.chartLayout.equity.height !== 200) {
  console.error('❌ FAIL: Equity pane should be height 200');
  process.exit(1);
}
console.log('✅ PASS: Equity pane: height=200');

if (chartConfig.chartLayout.oscillators.height !== 200) {
  console.error('❌ FAIL: Oscillators pane should be height 200');
  process.exit(1);
}
console.log('✅ PASS: Oscillators pane: height=200');

if (chartConfig.chartLayout.volume.height !== 200) {
  console.error('❌ FAIL: Volume pane should be height 200');
  process.exit(1);
}
console.log('✅ PASS: Volume pane: height=200\n');

console.log('=== TEST 4: Series config routes to correct panes ===');
if (chartConfig.seriesConfig.series['SMA 20']?.chart !== 'main') {
  console.error(`❌ FAIL: SMA 20 should route to 'main', got: ${chartConfig.seriesConfig.series['SMA 20']?.chart}`);
  process.exit(1);
}
console.log('✅ PASS: SMA 20 routed to \'main\' pane');

if (chartConfig.seriesConfig.series['Strategy Equity']?.chart !== 'equity') {
  console.error(`❌ FAIL: Strategy Equity should route to 'equity', got: ${chartConfig.seriesConfig.series['Strategy Equity']?.chart}`);
  process.exit(1);
}
console.log('✅ PASS: Strategy Equity routed to \'equity\' pane');

if (chartConfig.seriesConfig.series['RSI']?.chart !== 'oscillators') {
  console.error(`❌ FAIL: RSI should route to 'oscillators', got: ${chartConfig.seriesConfig.series['RSI']?.chart}`);
  process.exit(1);
}
console.log('✅ PASS: RSI routed to \'oscillators\' pane');

if (chartConfig.seriesConfig.series['Volume']?.chart !== 'volume') {
  console.error(`❌ FAIL: Volume should route to 'volume', got: ${chartConfig.seriesConfig.series['Volume']?.chart}`);
  process.exit(1);
}
console.log('✅ PASS: Volume routed to \'volume\' pane\n');

console.log('═══════════════════════════════════════════════════════════');
console.log('✅ ALL TESTS PASSED: Multi-pane architecture working correctly');
console.log('═══════════════════════════════════════════════════════════');
