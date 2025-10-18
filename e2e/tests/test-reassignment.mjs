#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { createProviderChain, DEFAULTS } from '../../src/config.js';
import { readFile } from 'fs/promises';

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

console.log('=== Reassignment ===\n');

const pineCode = await readFile('e2e/fixtures/strategies/test-reassignment.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);

const result = await runner.runPineScriptStrategy('BTCUSDT', '1h', 10, jsCode, 'test-reassignment.pine');

console.log('\n=== TEST RESULTS ===\n');

// Test 1: Simple Counter
const simpleCounter = result.plots?.['Simple Counter']?.data?.map(d => d.value) || [];
console.log('✓ Test 1 - Simple Counter (expected [1, 2, 3, ...]):');
console.log('  ', simpleCounter.slice(0, 5));
const test1Pass = simpleCounter.every((v, i) => v === i + 1);
console.log('  ', test1Pass ? '✅ PASS' : '❌ FAIL');

// Test 2: Step Counter
const stepCounter = result.plots?.['Step Counter +2']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 2 - Step Counter +2 (expected [2, 4, 6, ...]):');
console.log('  ', stepCounter.slice(0, 5));
const test2Pass = stepCounter.every((v, i) => v === (i + 1) * 2);
console.log('  ', test2Pass ? '✅ PASS' : '❌ FAIL');

// Test 3: Conditional Counter
const conditionalCounter = result.plots?.['Conditional Counter']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 3 - Conditional Counter (increments on close > close[1]):');
console.log('   First 5:', conditionalCounter.slice(0, 5));
console.log('   Length:', conditionalCounter.length);
const test3Pass = conditionalCounter.length === 10 && conditionalCounter.every((v, i) => !isNaN(v) && v !== null);
console.log('  ', test3Pass ? '✅ PASS' : '❌ FAIL');

// Test 4: Running High
const runningHigh = result.plots?.['Running High']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 4 - Running High (monotonically increasing):');
console.log('  ', runningHigh.slice(0, 5));
const test4Pass = runningHigh.every((v, i, arr) => i === 0 || v >= arr[i - 1]);
console.log('  ', test4Pass ? '✅ PASS' : '❌ FAIL');

// Test 5: Running Low
const runningLow = result.plots?.['Running Low']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 5 - Running Low (monotonically decreasing):');
console.log('  ', runningLow.slice(0, 5));
const test5Pass = runningLow.every((v, i, arr) => i === 0 || v <= arr[i - 1]);
console.log('  ', test5Pass ? '✅ PASS' : '❌ FAIL');

// Test 6: Trade State
const tradeState = result.plots?.['Trade State']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 6 - Trade State (0 or 100):');
console.log('  ', tradeState.slice(0, 5));
const test6Pass = tradeState.every(v => v === 0 || v === 100);
console.log('  ', test6Pass ? '✅ PASS' : '❌ FAIL');

// Test 7: Trailing Level
const trailingLevel = result.plots?.['Trailing Level']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 7 - Trailing Level (steps up by 10):');
console.log('  ', trailingLevel.slice(0, 5));
const test7Pass = trailingLevel.every((v, i, arr) => i === 0 || v >= arr[i - 1]);
console.log('  ', test7Pass ? '✅ PASS' : '❌ FAIL');

// Test 8: Multi-Historical
const multiHist = result.plots?.['Multi-Historical']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 8 - Multi-Historical ([1], [2], [3] references):');
console.log('  ', multiHist.slice(0, 5));
const test8Pass = multiHist.every(v => !isNaN(v) && v !== null);
console.log('  ', test8Pass ? '✅ PASS' : '❌ FAIL');

// Summary
const allTests = [test1Pass, test2Pass, test3Pass, test4Pass, test5Pass, 
                   test6Pass, test7Pass, test8Pass];
const passCount = allTests.filter(t => t).length;

console.log('\n=== SUMMARY ===');
console.log(`${passCount}/8 tests passed`);
console.log(passCount === 8 ? '✅ ALL TESTS PASS' : '❌ SOME TESTS FAILED');

process.exit(passCount === 8 ? 0 : 1);
