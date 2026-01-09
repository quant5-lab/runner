#!/usr/bin/env node
/* Strategy exit mechanisms E2E tests */
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test: Strategy Exit Mechanisms - Comprehensive');
console.log('═══════════════════════════════════════════════════════════\n');

const runTest = async (testName, pineFile, dataPattern, expectedBehavior) => {
  console.log(`\n─────────────────────────────────────────────────────────`);
  console.log(`Test: ${testName}`);
  console.log(`─────────────────────────────────────────────────────────`);

  const mockProvider = new MockProviderManager({
    dataPattern,
    basePrice: 100,
    amplitude: 10,
  });

  const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
  const DEFAULTS = { showDebug: false, showStats: false };

  const container = createContainer(createProviderChain, DEFAULTS);
  const runner = container.resolve('tradingAnalysisRunner');
  const transpiler = container.resolve('pineScriptTranspiler');

  const pineCode = await readFile(pineFile, 'utf-8');
  const jsCode = await transpiler.transpile(pineCode);

  const result = await runner.runPineScriptStrategy('TEST', '1h', 100, jsCode, pineFile);

  const getVals = (title) =>
    result.plots?.[title]?.data?.map((d) => d.value).filter((v) => v != null) || [];

  const posSize = getVals('Position Size');
  const closedTrades = result.closedTrades || 0;
  const openTrades = result.openTrades || 0;

  console.log('\n=== EXECUTION RESULTS ===');
  console.log('Closed trades:      ', closedTrades);
  console.log('Open trades:        ', openTrades);
  console.log('Position samples:   ', posSize.slice(50, 60));

  const passed = expectedBehavior(closedTrades, openTrades, posSize);
  console.log('\n' + (passed ? '✅ PASS' : '❌ FAIL') + ': ' + testName);

  return passed;
};

const test1 = await runTest(
  'Immediate Exit - close_all() on single condition',
  'e2e/fixtures/strategies/test-exit-immediate.pine',
  'sawtooth',
  (closed, open, pos) => {
    return closed > 0 && open === 0;
  },
);

const test2 = await runTest(
  'Delayed Exit - state[N] historical reference logic',
  'e2e/fixtures/strategies/test-exit-delayed-state.pine',
  'sawtooth',
  (closed, open, pos) => {
    return closed > 0;
  },
);

const test3 = await runTest(
  'Selective Exit - strategy.close(id) vs close_all()',
  'e2e/fixtures/strategies/test-exit-selective.pine',
  'sawtooth',
  (closed, open, pos) => {
    return closed > 0;
  },
);

const test4 = await runTest(
  'Complex Exit - multi-bar condition evaluation',
  'e2e/fixtures/strategies/test-exit-multibar-condition.pine',
  'sawtooth',
  (closed, open, pos) => {
    return closed > 0;
  },
);

const test5 = await runTest(
  'Exit Reset - condition state after position close',
  'e2e/fixtures/strategies/test-exit-state-reset.pine',
  'bullish',
  (closed, open, pos) => {
    return closed >= 2;
  },
);

console.log('\n═══════════════════════════════════════════════════════════');
console.log('SUMMARY');
console.log('═══════════════════════════════════════════════════════════');
const allTests = [test1, test2, test3, test4, test5];
const passed = allTests.filter((t) => t).length;
const total = allTests.length;

console.log(`Tests passed: ${passed}/${total}`);

if (passed === total) {
  console.log('✅ ALL TESTS PASSED\n');
  process.exit(0);
} else {
  console.log('❌ SOME TESTS FAILED\n');
  process.exit(1);
}
