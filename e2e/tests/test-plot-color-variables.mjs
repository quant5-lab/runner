#!/usr/bin/env node
/**
 * E2E Test: Variable References in Plot Color Expressions
 *
 * Validates that variables used in plot color expressions are correctly
 * transpiled and executed. Tests various patterns of variable usage in
 * color parameters including simple variables, strategy properties, and
 * complex expressions.
 */

import { createContainer } from '../../src/container.js';
import { MockProviderManager } from '../mocks/MockProvider.js';

let passed = 0;
let failed = 0;

function assert(condition, message) {
  if (!condition) {
    console.error(`❌ FAIL: ${message}`);
    failed++;
    throw new Error(message);
  }
  console.log(`✅ PASS: ${message}`);
  passed++;
}

async function runTest(testName, pineCode) {
  console.log(`\n🧪 ${testName}`);
  console.log('='.repeat(80));

  try {
    const mockProvider = new MockProviderManager({
      dataPattern: 'linear',
      basePrice: 100,
      amplitude: 10,
    });
    const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
    const DEFAULTS = { showDebug: false, showStats: false };

    const container = createContainer(createProviderChain, DEFAULTS);
    const runner = container.resolve('tradingAnalysisRunner');
    const transpiler = container.resolve('pineScriptTranspiler');

    const jsCode = await transpiler.transpile(pineCode);
    const result = await runner.runPineScriptStrategy('TEST', '1h', 10, jsCode, 'inline-test.pine');

    // Validate execution
    assert(result, 'Strategy executed without error');
    assert(result.plots, 'Has plots object');
    assert(Object.keys(result.plots).length > 0, 'Generated at least one plot');

    console.log(`✅ ${testName} PASSED`);
  } catch (error) {
    console.error(`❌ ${testName} FAILED:`, error.message);
    failed++;
  }
}

console.log('Testing comprehensive variable usage in plot color expressions...\n');

// ============================================================================
// TEST 1: Simple variable in color expression
// ============================================================================
await runTest(
  'Simple variable in plot color',
  `//@version=4
strategy("Test 1", overlay=true)

has_active = close > open
plot(close, color=has_active ? color.green : color.red)
`,
);

// ============================================================================
// TEST 2: Variable with strategy.position_avg_price
// ============================================================================
await runTest(
  'Variable with strategy.position_avg_price',
  `//@version=4
strategy("Test 2", overlay=true)

has_position = not na(strategy.position_avg_price)
plot(close, color=has_position ? color.blue : color.gray)
`,
);

// ============================================================================
// TEST 3: Multiple variables in color expression
// ============================================================================
await runTest(
  'Multiple variables in color',
  `//@version=4
strategy("Test 3", overlay=true)

bullish = close > open
strong = volume > volume[1]
plot(close, color=bullish and strong ? color.green : color.red)
`,
);

// ============================================================================
// TEST 4: Nested conditional with variables in color
// ============================================================================
await runTest(
  'Nested conditional with variables in color',
  `//@version=4
strategy("Test Nested Color", overlay=true)

up = close > open
strong_up = up and (volume > volume[1])
weak_up = up and (volume <= volume[1])

color_val = strong_up ? color.green : weak_up ? color.lime : color.red

plot(close, color=color_val)
`,
);

// ============================================================================
// TEST 5: Variable used in multiple plot parameters
// ============================================================================
await runTest(
  'Variable in multiple plot parameters',
  `//@version=4
strategy("Test Multi Param", overlay=true)

is_bullish = close > open
line_width = is_bullish ? 3 : 1

plot(close, color=is_bullish ? color.green : color.red, linewidth=line_width)
`,
);

// ============================================================================
// TEST 6: Variable with has_active_trade pattern
// ============================================================================
await runTest(
  'Variable with has_active_trade pattern',
  `//@version=4
strategy("Test Has Active Trade", overlay=true)

has_active_trade = not na(strategy.position_avg_price)
stop_level = close * 0.95

plot(stop_level, color=has_active_trade ? color.red : color.white)
`,
);

// ============================================================================
// SUMMARY
// ============================================================================
console.log('\n' + '='.repeat(80));
console.log('TEST SUMMARY');
console.log('='.repeat(80));
console.log(`✅ Tests Passed: ${passed}`);
console.log(`❌ Tests Failed: ${failed}`);

if (failed > 0) {
  console.log('\n❌ SOME TESTS FAILED');
  process.exit(1);
} else {
  console.log('\n✅ ALL TESTS PASSED');
  process.exit(0);
}
