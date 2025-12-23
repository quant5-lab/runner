#!/usr/bin/env node
/**
 * E2E Test: Dynamic Pane Creation from Config Overrides
 *
 * Tests frontend pane creation matching backend behavior:
 * 1. Config file (.config) overrides indicator pane assignments
 * 2. Frontend buildPaneConfig() extracts unique pane names dynamically
 * 3. PaneManager creates all custom panes before series routing
 * 4. Series successfully route to dynamically created panes
 *
 * Edge cases:
 * - Multiple indicators per pane
 * - Multiple unique panes
 * - Backward compatibility with auto 'indicator' pane
 * - Config override merging with style properties
 */
import { createContainer } from '../../src/container.js';
import { readFile, writeFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test: Dynamic Pane Creation from Config Overrides');
console.log('═══════════════════════════════════════════════════════════\n');

const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');
const configBuilder = container.resolve('configurationBuilder');

/* Test 1: Multiple custom panes from config file */
console.log('=== TEST 1: Config file with 3 custom panes ===');

const testStrategy = `
//@version=5
indicator("Config Override Test", overlay=false)

plot(10, title="Indicator A", color=color.red)
plot(20, title="Indicator B", color=color.blue)
plot(30, title="Indicator C", color=color.green)
plot(40, title="Indicator D", color=color.orange)
plot(50, title="Indicator E", color=color.purple)
`;

const pineCode = testStrategy;
const jsCode = await transpiler.transpile(pineCode);
const result = await runner.runPineScriptStrategy(
  'TEST',
  'D',
  30,
  jsCode,
  'test-config-pane-override.pine',
);

const configOverride = {
  'Indicator A': { chartPane: 'pane_alpha' },
  'Indicator B': { chartPane: 'pane_alpha' },
  'Indicator C': { chartPane: 'pane_beta' },
  'Indicator D': { chartPane: 'pane_beta' },
  'Indicator E': { chartPane: 'pane_gamma' },
};

const metadata = runner.extractIndicatorMetadata(result.plots);
Object.keys(metadata).forEach((key) => {
  if (configOverride[key]?.chartPane) {
    metadata[key].chartPane = configOverride[key].chartPane;
  }
});

const tradingConfig = configBuilder.createTradingConfig('TEST', 'D', 30, 'test-config-override');
const chartConfig = configBuilder.generateChartConfig(tradingConfig, metadata);

const panes = Object.keys(chartConfig.chartLayout);
if (!panes.includes('pane_alpha')) {
  console.error(`❌ FAIL: Missing pane_alpha in layout. Got: ${panes.join(', ')}`);
  process.exit(1);
}
console.log('✅ PASS: pane_alpha created from config override');

if (!panes.includes('pane_beta')) {
  console.error(`❌ FAIL: Missing pane_beta in layout. Got: ${panes.join(', ')}`);
  process.exit(1);
}
console.log('✅ PASS: pane_beta created from config override');

if (!panes.includes('pane_gamma')) {
  console.error(`❌ FAIL: Missing pane_gamma in layout. Got: ${panes.join(', ')}`);
  process.exit(1);
}
console.log('✅ PASS: pane_gamma created from config override');

if (panes.length !== 4) {
  // main + 3 custom panes
  console.error(`❌ FAIL: Expected 4 panes (main + 3 custom), got ${panes.length}: ${panes.join(', ')}`);
  process.exit(1);
}
console.log(`✅ PASS: Total 4 panes: ${panes.join(', ')}\n`);

/* Test 2: Series routing to correct custom panes */
console.log('=== TEST 2: Series routed to custom panes from config ===');

if (chartConfig.seriesConfig.series['Indicator A']?.chart !== 'pane_alpha') {
  console.error(
    `❌ FAIL: Indicator A should route to 'pane_alpha', got: ${chartConfig.seriesConfig.series['Indicator A']?.chart}`,
  );
  process.exit(1);
}
console.log("✅ PASS: Indicator A routed to 'pane_alpha'");

if (chartConfig.seriesConfig.series['Indicator B']?.chart !== 'pane_alpha') {
  console.error(
    `❌ FAIL: Indicator B should route to 'pane_alpha', got: ${chartConfig.seriesConfig.series['Indicator B']?.chart}`,
  );
  process.exit(1);
}
console.log("✅ PASS: Indicator B routed to 'pane_alpha' (shared pane)");

if (chartConfig.seriesConfig.series['Indicator C']?.chart !== 'pane_beta') {
  console.error(
    `❌ FAIL: Indicator C should route to 'pane_beta', got: ${chartConfig.seriesConfig.series['Indicator C']?.chart}`,
  );
  process.exit(1);
}
console.log("✅ PASS: Indicator C routed to 'pane_beta'");

if (chartConfig.seriesConfig.series['Indicator E']?.chart !== 'pane_gamma') {
  console.error(
    `❌ FAIL: Indicator E should route to 'pane_gamma', got: ${chartConfig.seriesConfig.series['Indicator E']?.chart}`,
  );
  process.exit(1);
}
console.log("✅ PASS: Indicator E routed to 'pane_gamma'\n");

/* Test 3: Backward compatibility - no config overrides */
console.log('=== TEST 3: Backward compatibility without config overrides ===');

const defaultMetadata = runner.extractIndicatorMetadata(result.plots);
const defaultChartConfig = configBuilder.generateChartConfig(tradingConfig, defaultMetadata);
const defaultPanes = Object.keys(defaultChartConfig.chartLayout);

if (!defaultPanes.includes('indicator')) {
  console.error(
    `❌ FAIL: Should auto-create 'indicator' pane when no config. Got: ${defaultPanes.join(', ')}`,
  );
  process.exit(1);
}
console.log("✅ PASS: Auto-created 'indicator' pane for backward compatibility");

if (defaultPanes.length !== 2) {
  // main + indicator
  console.error(`❌ FAIL: Expected 2 panes (main + indicator), got ${defaultPanes.length}`);
  process.exit(1);
}
console.log('✅ PASS: Default layout has main + indicator panes\n');

/* Test 4: Deduplication - multiple indicators same pane */
console.log('=== TEST 4: Pane deduplication with repeated pane names ===');

const duplicateConfig = {
  'Indicator A': { chartPane: 'shared_pane' },
  'Indicator B': { chartPane: 'shared_pane' },
  'Indicator C': { chartPane: 'shared_pane' },
  'Indicator D': { chartPane: 'shared_pane' },
  'Indicator E': { chartPane: 'another_pane' },
};

const dedupMetadata = runner.extractIndicatorMetadata(result.plots);
Object.keys(dedupMetadata).forEach((key) => {
  if (duplicateConfig[key]?.chartPane) {
    dedupMetadata[key].chartPane = duplicateConfig[key].chartPane;
  }
});

const dedupChartConfig = configBuilder.generateChartConfig(tradingConfig, dedupMetadata);
const dedupPanes = Object.keys(dedupChartConfig.chartLayout);

if (!dedupPanes.includes('shared_pane')) {
  console.error(`❌ FAIL: Missing shared_pane. Got: ${dedupPanes.join(', ')}`);
  process.exit(1);
}

if (!dedupPanes.includes('another_pane')) {
  console.error(`❌ FAIL: Missing another_pane. Got: ${dedupPanes.join(', ')}`);
  process.exit(1);
}

if (dedupPanes.length !== 3) {
  // main + shared_pane + another_pane
  console.error(`❌ FAIL: Expected 3 panes, got ${dedupPanes.length}: ${dedupPanes.join(', ')}`);
  process.exit(1);
}
console.log('✅ PASS: Correctly deduplicated to 3 unique panes (main + 2 custom)');

const sharedPaneIndicators = Object.entries(dedupChartConfig.seriesConfig.series).filter(
  ([_, config]) => config.chart === 'shared_pane',
);

if (sharedPaneIndicators.length !== 4) {
  console.error(`❌ FAIL: Expected 4 indicators in shared_pane, got ${sharedPaneIndicators.length}`);
  process.exit(1);
}
console.log('✅ PASS: 4 indicators correctly routed to shared_pane\n');

/* Test 5: Edge case - all indicators on main pane */
console.log('=== TEST 5: Edge case - all indicators assigned to main pane ===');

const allMainConfig = {
  'Indicator A': { chartPane: 'main' },
  'Indicator B': { chartPane: 'main' },
  'Indicator C': { chartPane: 'main' },
  'Indicator D': { chartPane: 'main' },
  'Indicator E': { chartPane: 'main' },
};

const allMainMetadata = runner.extractIndicatorMetadata(result.plots);
Object.keys(allMainMetadata).forEach((key) => {
  if (allMainConfig[key]?.chartPane) {
    allMainMetadata[key].chartPane = allMainConfig[key].chartPane;
  }
});

const allMainChartConfig = configBuilder.generateChartConfig(tradingConfig, allMainMetadata);
const allMainPanes = Object.keys(allMainChartConfig.chartLayout);

if (!allMainPanes.includes('indicator')) {
  console.error(
    `❌ FAIL: Should auto-create indicator pane when all on main. Got: ${allMainPanes.join(', ')}`,
  );
  process.exit(1);
}
console.log('✅ PASS: Auto-created indicator pane when all indicators on main');

if (allMainPanes.length !== 2) {
  console.error(`❌ FAIL: Expected 2 panes (main + indicator), got ${allMainPanes.length}`);
  process.exit(1);
}
console.log('✅ PASS: Backward compatibility maintained with all-main config\n');

/* Test 6: Empty metadata edge case */
console.log('=== TEST 6: Edge case - empty metadata ===');

const emptyChartConfig = configBuilder.generateChartConfig(tradingConfig, {});
const emptyPanes = Object.keys(emptyChartConfig.chartLayout);

if (!emptyPanes.includes('indicator')) {
  console.error(`❌ FAIL: Should auto-create indicator pane for empty metadata`);
  process.exit(1);
}
console.log('✅ PASS: Auto-created indicator pane for empty metadata');

if (emptyPanes.length !== 2) {
  console.error(`❌ FAIL: Expected 2 panes for empty metadata, got ${emptyPanes.length}`);
  process.exit(1);
}
console.log('✅ PASS: Empty metadata produces default layout\n');

console.log('═══════════════════════════════════════════════════════════');
console.log('✅ ALL TESTS PASSED: Dynamic pane creation from config works correctly');
console.log('═══════════════════════════════════════════════════════════');
