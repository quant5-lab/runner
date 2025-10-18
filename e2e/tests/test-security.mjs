#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { createProviderChain, DEFAULTS } from '../../src/config.js';
import { readFile } from 'fs/promises';

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

console.log('=== Security Function ===\n');

const pineCode = await readFile('e2e/fixtures/strategies/test-security.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);

const result = await runner.runPineScriptStrategy('SBER', '1h', 50, jsCode, 'test-security.pine');

console.log('\n=== TEST RESULTS ===\n');

// Test 1: SMA20 Daily - should have varying values
const sma20Daily = result.plots?.['SMA20 Daily']?.data?.map(d => d.value) || [];
console.log('✓ Test 1 - SMA20 Daily values:');
console.log('   First 5:', sma20Daily.slice(0, 5));
console.log('   Last 5:', sma20Daily.slice(-5));

// Check that values are not all identical (security is working)
const uniqueValues = new Set(sma20Daily);
const test1Pass = uniqueValues.size > 1 && sma20Daily.every(v => !isNaN(v) && v !== null && v > 0);
console.log('   Unique values:', uniqueValues.size);
console.log('  ', test1Pass ? '✅ PASS' : '❌ FAIL - All values identical or invalid');

// Test 2: Daily Close - should have varying values
const dailyClose = result.plots?.['Daily Close']?.data?.map(d => d.value) || [];
console.log('\n✓ Test 2 - Daily Close values:');
console.log('   First 5:', dailyClose.slice(0, 5));
console.log('   Last 5:', dailyClose.slice(-5));

const uniqueCloseValues = new Set(dailyClose);
const test2Pass = uniqueCloseValues.size > 1 && dailyClose.every(v => !isNaN(v) && v !== null && v > 0);
console.log('   Unique values:', uniqueCloseValues.size);
console.log('  ', test2Pass ? '✅ PASS' : '❌ FAIL - All values identical or invalid');

// Test 3: Values should differ between SMA and Close
const test3Pass = sma20Daily.some((v, i) => Math.abs(v - dailyClose[i]) > 0.01);
console.log('\n✓ Test 3 - SMA differs from Close:');
console.log('  ', test3Pass ? '✅ PASS' : '❌ FAIL - SMA equals Close everywhere');

// Summary
const allTests = [test1Pass, test2Pass, test3Pass];
const passCount = allTests.filter(t => t).length;

console.log('\n=== SUMMARY ===');
console.log(`${passCount}/3 tests passed`);
console.log(passCount === 3 ? '✅ ALL TESTS PASS' : '❌ SOME TESTS FAILED');

process.exit(passCount === 3 ? 0 : 1);
