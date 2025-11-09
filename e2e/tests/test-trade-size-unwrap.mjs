#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';

/* Test: Validate trade size is numeric, not param() wrapped object */
console.log('═══════════════════════════════════════════════════════════');
console.log('TEST: Trade Size Unwrap Validation');
console.log('═══════════════════════════════════════════════════════════\n');

const testPattern = 'linear'; // Predictable data for testing
const bars = 50;

/* Setup test environment */
const mockProvider = new MockProviderManager({ dataPattern: testPattern });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

/* Transpile and execute strategy */
const strategyCode = await readFile('e2e/fixtures/strategies/test-trade-size-unwrap.pine', 'utf8');
const jsCode = await transpiler.transpile(strategyCode);
const result = await runner.runPineScriptStrategy(
  'TEST',
  '1h',
  bars,
  jsCode,
  'test-trade-size-unwrap.pine',
);

console.log('=== STRATEGY EXECUTION ===');
console.log('Closed trades:       ', result.strategy?.trades?.length || 0);
console.log('Open trades:         ', result.strategy?.openTrades?.length || 0);

/* Validate trade data structure */
if (!result.strategy?.trades || result.strategy.trades.length === 0) {
  console.error('\n❌ FAILED: No trades found');
  console.error('Expected: At least 1 closed trade from test strategy');
  console.error('Actual:   0 trades');
  process.exit(1);
}

const trades = result.strategy.trades;
console.log('\n=== TRADE SIZE VALIDATION ===');

let allTestsPassed = true;
const validationResults = [];

for (let i = 0; i < trades.length; i++) {
  const trade = trades[i];
  const testResult = {
    tradeNum: i + 1,
    entryId: trade.entryId,
    sizeType: typeof trade.size,
    sizeValue: trade.size,
    isNumeric: typeof trade.size === 'number',
    isValidNumber: typeof trade.size === 'number' && !isNaN(trade.size) && isFinite(trade.size),
    profitCalculated: trade.profit !== null && trade.profit !== undefined,
    profitValue: trade.profit,
  };

  validationResults.push(testResult);

  /* Test 1: Size must be numeric type */
  if (!testResult.isNumeric) {
    console.error(`\n❌ Trade #${i + 1} (${trade.entryId}): Size is not numeric`);
    console.error('   Expected: number');
    console.error(`   Actual:   ${testResult.sizeType}`);
    console.error(`   Value:    ${JSON.stringify(trade.size).substring(0, 200)}...`);
    allTestsPassed = false;
  }

  /* Test 2: Size must be valid number (not NaN, not Infinity) */
  if (testResult.isNumeric && !testResult.isValidNumber) {
    console.error(`\n❌ Trade #${i + 1} (${trade.entryId}): Size is invalid number`);
    console.error(`   Value: ${trade.size}`);
    allTestsPassed = false;
  }

  /* Test 3: Profit must be calculated (not null) */
  if (!testResult.profitCalculated) {
    console.error(`\n❌ Trade #${i + 1} (${trade.entryId}): Profit not calculated`);
    console.error('   Expected: numeric profit value');
    console.error(`   Actual:   ${trade.profit}`);
    allTestsPassed = false;
  }

  /* Test 4: Size must be positive */
  if (testResult.isValidNumber && trade.size <= 0) {
    console.error(`\n❌ Trade #${i + 1} (${trade.entryId}): Size must be positive`);
    console.error(`   Value: ${trade.size}`);
    allTestsPassed = false;
  }
}

/* Display validation summary */
console.log('\n=== VALIDATION SUMMARY ===');
for (const result of validationResults) {
  const sizeStatus = result.isValidNumber ? '✅' : '❌';
  const profitStatus = result.profitCalculated ? '✅' : '❌';

  console.log(`Trade #${result.tradeNum} (${result.entryId}):`);
  console.log(`  ${sizeStatus} Size: ${result.sizeType} = ${result.sizeValue}`);
  console.log(`  ${profitStatus} Profit: ${result.profitValue?.toFixed(2) || 'null'}`);
}

/* Test edge cases specific to unwrapping */
console.log('\n=== UNWRAP EDGE CASES ===');

for (const trade of trades) {
  /* Test: Size should not have param() structure markers */
  if (typeof trade.size === 'object' && trade.size !== null) {
    if ('when' in trade.size || 'else' in trade.size) {
      console.error(`❌ Trade ${trade.entryId}: Size contains param() structure`);
      console.error(`   Size object keys: ${Object.keys(trade.size)}`);
      allTestsPassed = false;
    }
  }

  /* Test: Entry/exit prices should be numeric */
  if (typeof trade.entryPrice !== 'number' || isNaN(trade.entryPrice)) {
    console.error(`❌ Trade ${trade.entryId}: Invalid entry price: ${trade.entryPrice}`);
    allTestsPassed = false;
  }

  if (trade.exitPrice && (typeof trade.exitPrice !== 'number' || isNaN(trade.exitPrice))) {
    console.error(`❌ Trade ${trade.entryId}: Invalid exit price: ${trade.exitPrice}`);
    allTestsPassed = false;
  }

  /* Test: Bar indices should be integers */
  if (!Number.isInteger(trade.entryBar)) {
    console.error(`❌ Trade ${trade.entryId}: Entry bar not integer: ${trade.entryBar}`);
    allTestsPassed = false;
  }
}

if (allTestsPassed) {
  console.log('✅ All edge case validations passed');
}

console.log('\n═══════════════════════════════════════════════════════════');
console.log('RESULTS');
console.log('═══════════════════════════════════════════════════════════');

if (allTestsPassed) {
  console.log('✅ ALL TESTS PASSED');
  console.log('✅ Trade size field correctly unwrapped to numeric values');
  console.log('✅ Profit calculations working correctly');
  console.log(`✅ ${trades.length} trade(s) validated successfully`);
  process.exit(0);
} else {
  console.log('❌ SOME TESTS FAILED');
  console.log('❌ Trade size unwrapping issue detected');
  console.log('⚠️  This indicates param() objects are not being fully unwrapped');
  console.log('⚠️  See PineTS docs/ISSUE_TRADE_SIZE_UNWRAP.md for details');
  process.exit(1);
}
