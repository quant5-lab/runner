#!/usr/bin/env node
/* bar_index built-in variable E2E tests */
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test: bar_index Built-in Variable - Comprehensive');
console.log('═══════════════════════════════════════════════════════════\n');

const runTest = async (testName, pineFile, expectedBehavior) => {
  console.log(`\n─────────────────────────────────────────────────────────`);
  console.log(`Test: ${testName}`);
  console.log(`─────────────────────────────────────────────────────────`);

  const mockProvider = new MockProviderManager({
    dataPattern: 'linear',
    basePrice: 100,
    amplitude: 1,
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
    result.plots?.[title]?.data?.map((d) => d.value).filter((v) => v != null && !isNaN(v)) || [];

  const passed = expectedBehavior(result, getVals);
  console.log('\n' + (passed ? '✅ PASS' : '❌ FAIL') + ': ' + testName);

  return passed;
};

const test1 = await runTest(
  'Basic bar_index - main context sequential values',
  'e2e/fixtures/strategies/test-bar-index-basic.pine',
  (result, getVals) => {
    const barIndexVals = getVals('Bar Index');
    console.log('Bar index values (first 10):', barIndexVals.slice(0, 10));
    console.log('Bar index values (last 5):', barIndexVals.slice(-5));

    if (barIndexVals.length < 10) return false;
    if (barIndexVals[0] !== 0) return false;
    if (barIndexVals[1] !== 1) return false;
    if (barIndexVals[9] !== 9) return false;

    for (let i = 0; i < barIndexVals.length; i++) {
      if (barIndexVals[i] !== i) {
        console.log(`  ❌ Sequence break at index ${i}: expected ${i}, got ${barIndexVals[i]}`);
        return false;
      }
    }

    return true;
  },
);

const test2 = await runTest(
  'Modulo Operations - bar_index % N patterns',
  'e2e/fixtures/strategies/test-bar-index-modulo.pine',
  (result, getVals) => {
    const mod5 = getVals('Mod 5');
    const mod10 = getVals('Mod 10');
    const mod20 = getVals('Mod 20');

    console.log('Mod 5 pattern (first 15):', mod5.slice(0, 15));
    console.log('Mod 20 pattern (first 25):', mod20.slice(0, 25));

    if (mod5[0] !== 0 || mod5[5] !== 0 || mod5[10] !== 0) return false;
    if (mod5[3] !== 3 || mod5[8] !== 3) return false;

    if (mod20[0] !== 0 || mod20[20] !== 0 || mod20[40] !== 0) return false;

    return true;
  },
);

const test3 = await runTest(
  'Security Context - bar_index in security() (bb9 pattern)',
  'e2e/fixtures/strategies/test-bar-index-security.pine',
  (result, getVals) => {
    const secBarIndex = getVals('Security Bar Index');
    const secMod20 = getVals('Security Mod 20');

    console.log('Security bar_index (first 10):', secBarIndex.slice(0, 10));
    console.log('Security mod 20 (bars 0-50):', secMod20.slice(0, 51));

    const hasNaN = secBarIndex.some((v) => isNaN(v));
    if (hasNaN) {
      console.log('  ❌ CRITICAL: NaN values in security bar_index');
      return false;
    }

    if (secMod20.length === 0) {
      console.log('  ❌ CRITICAL: No mod 20 values from security()');
      return false;
    }

    return true;
  },
);

const test4 = await runTest(
  'Conditional Logic - bar_index in if/ternary statements',
  'e2e/fixtures/strategies/test-bar-index-conditional.pine',
  (result, getVals) => {
    const firstBar = getVals('First Bar Flag');
    const everyTenBars = getVals('Every 10 Bars');

    console.log('First bar flag:', firstBar[0], firstBar[1]);
    console.log('Every 10 bars (0-30):', everyTenBars.slice(0, 31));

    if (firstBar[0] !== 1) return false;
    if (firstBar[1] !== 0) return false;

    if (everyTenBars[0] !== 1 || everyTenBars[10] !== 1 || everyTenBars[20] !== 1) return false;
    if (everyTenBars[5] !== 0 || everyTenBars[15] !== 0) return false;

    return true;
  },
);

const test5 = await runTest(
  'Comparisons - bar_index > N, < N, == N',
  'e2e/fixtures/strategies/test-bar-index-comparisons.pine',
  (result, getVals) => {
    const gtTen = getVals('Greater Than 10');
    const eqTwenty = getVals('Equals 20');
    const rangeCheck = getVals('In Range');

    console.log('Greater than 10 (bars 8-12):', gtTen.slice(8, 13));
    console.log('Equals 20 (bars 18-22):', eqTwenty.slice(18, 23));

    if (gtTen[10] !== 0 || gtTen[11] !== 1) return false;

    if (eqTwenty[19] !== 0 || eqTwenty[20] !== 1 || eqTwenty[21] !== 0) return false;

    return true;
  },
);

const test6 = await runTest(
  'Historical Access - bar_index[N] lookback',
  'e2e/fixtures/strategies/test-bar-index-historical.pine',
  (result, getVals) => {
    const prevBar = getVals('Previous Bar Index');
    const barDiff = getVals('Bar Difference');

    console.log('Previous bar index (bars 0-5):', prevBar.slice(0, 6));
    console.log('Bar difference (bars 0-5):', barDiff.slice(0, 6));

    if (prevBar.length > 5) {
      if (prevBar[5] !== 4) return false;
    }

    if (barDiff.length > 5) {
      if (barDiff[5] !== 1) return false;
    }

    return true;
  },
);

console.log('\n═══════════════════════════════════════════════════════════');
console.log('SUMMARY');
console.log('═══════════════════════════════════════════════════════════');
const allTests = [test1, test2, test3, test4, test5, test6];
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
