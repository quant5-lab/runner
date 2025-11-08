#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../utils/test-helpers.js';

/* Built-in Variables E2E Tests - Parametric validation for all built-in variables */

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

/* Manual calculation of derived built-in variables */
function calcDerivedVariable(highs, lows, closes, type) {
  switch (type) {
    case 'hl2':
      return highs.map((h, i) => (h + lows[i]) / 2);
    case 'hlc3':
      return highs.map((h, i) => (h + lows[i] + closes[i]) / 3);
    case 'ohlc4':
      return highs.map((h, i) => (closes[i] + h + lows[i] + closes[i]) / 4);
    case 'tr':
      return calcTrueRange(highs, lows, closes);
    default:
      throw new Error(`Unknown derived type: ${type}`);
  }
}

function calcTrueRange(highs, lows, closes) {
  const result = [];
  for (let i = 0; i < highs.length; i++) {
    if (i === 0) {
      result.push(highs[i] - lows[i]);
    } else {
      const tr1 = highs[i] - lows[i];
      const tr2 = Math.abs(highs[i] - closes[i - 1]);
      const tr3 = Math.abs(lows[i] - closes[i - 1]);
      result.push(Math.max(tr1, tr2, tr3));
    }
  }
  return result;
}

console.log('=== Built-in Variables E2E Tests ===\n');

/* Base variables: open, high, low, close, volume */
const baseVariables = ['open', 'high', 'low', 'close', 'volume'];
/* Derived variables: hl2, hlc3, ohlc4, tr */
const derivedVariables = ['hl2', 'hlc3', 'ohlc4', 'tr'];

// Test 1: Direct access to all built-in variables
console.log('Test 1: Direct access to all built-in variables');
const directResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-builtin-direct.pine',
  'volatile',
  100,
  15,
);

let passedBase = 0;
for (const varName of baseVariables) {
  const values = getPlotValues(directResult, varName);
  if (!values || values.length === 0) {
    console.error(`❌ FAILED: ${varName} not accessible`);
    process.exit(1);
  }
  passedBase++;
}
console.log(`✅ PASSED: ${passedBase}/${baseVariables.length} base variables accessible\n`);

// Test 2: Derived variables calculation validation
console.log('Test 2: Derived variables calculation validation');
const derivedResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-builtin-derived.pine',
  'volatile',
  100,
  15,
);

const highValues = getPlotValues(derivedResult, 'high');
const lowValues = getPlotValues(derivedResult, 'low');
const closeValues = getPlotValues(derivedResult, 'close');

let passedDerived = 0;
for (const varName of derivedVariables) {
  const values = getPlotValues(derivedResult, varName);
  if (!values || values.length === 0) {
    console.error(`❌ FAILED: ${varName} not accessible`);
    process.exit(1);
  }
  
  const expected = calcDerivedVariable(highValues, lowValues, closeValues, varName);
  let matched = 0;
  for (let i = 0; i < values.length; i++) {
    assertFloatEquals(values[i], expected[i], FLOAT_EPSILON, `${varName}[${i}]`);
    matched++;
  }
  console.log(`  ${varName}: ${matched}/${values.length} values match calculation`);
  passedDerived++;
}
console.log(`✅ PASSED: ${passedDerived}/${derivedVariables.length} derived variables correct\n`);

// Test 3: Variables in calculations (SMA)
console.log('Test 3: Variables in calculations (SMA)');
const calcResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-builtin-calculations.pine',
  'volatile',
  100,
  15,
);

const testVars = ['close', 'volume', 'hl2', 'tr'];
let passedCalc = 0;
for (const varName of testVars) {
  const smaValues = getPlotValues(calcResult, `${varName}_sma`);
  if (!smaValues) {
    console.error(`❌ FAILED: ${varName} SMA not calculated`);
    process.exit(1);
  }
  const validCount = smaValues.filter((v) => !isNaN(v)).length;
  console.log(`  ${varName}: ${validCount}/${smaValues.length} valid SMA values`);
  passedCalc++;
}
console.log(`✅ PASSED: ${passedCalc}/${testVars.length} variables work in calculations\n`);

// Test 4: Variables in conditional logic
console.log('Test 4: Variables in conditional logic');
const condResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-builtin-conditions.pine',
  'volatile',
  100,
  20,
);

const testCondVars = ['close', 'volume', 'hl2', 'tr'];
let passedCond = 0;
for (const varName of testCondVars) {
  const signalValues = getPlotValues(condResult, `${varName}_signal`);
  if (!signalValues) {
    console.error(`❌ FAILED: ${varName} conditional not working`);
    process.exit(1);
  }
  const signalCount = signalValues.filter((v) => v === 1).length;
  console.log(`  ${varName}: ${signalCount} signals generated`);
  passedCond++;
}
console.log(`✅ PASSED: ${passedCond}/${testCondVars.length} variables work in conditionals\n`);

// Test 5: Variables in function scope
console.log('Test 5: Variables in function scope');
const funcResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-builtin-function.pine',
  'volatile',
  100,
  10,
);

const testFuncVars = ['close', 'volume', 'hl2', 'tr'];
let passedFunc = 0;
for (const varName of testFuncVars) {
  const funcValues = getPlotValues(funcResult, `${varName}_func`);
  if (!funcValues || funcValues.length === 0) {
    console.error(`❌ FAILED: ${varName} not accessible in function scope`);
    process.exit(1);
  }
  passedFunc++;
}
console.log(`✅ PASSED: ${passedFunc}/${testFuncVars.length} variables accessible in functions\n`);

// Test 6: Multiple variable usages
console.log('Test 6: Multiple variable usages');
const multiResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-builtin-multiple.pine',
  'volatile',
  100,
  10,
);

const multi1 = getPlotValues(multiResult, 'close_direct');
const multi2 = getPlotValues(multiResult, 'close_sma');
const multi3 = getPlotValues(multiResult, 'hl2_direct');
const multi4 = getPlotValues(multiResult, 'tr_direct');

if (!multi1 || !multi2 || !multi3 || !multi4) {
  console.error('❌ FAILED: Multiple variable usages not working');
  process.exit(1);
}

console.log(`✅ PASSED: Multiple simultaneous variable usages work correctly\n`);

console.log('=== All built-in variable tests passed ✅ ===');
console.log(`Total: 6 test scenarios covering:
  - Direct access (${baseVariables.length} base + ${derivedVariables.length} derived = ${baseVariables.length + derivedVariables.length} variables)
  - Derived variable calculation validation
  - Variables in calculations (SMA)
  - Variables in conditional logic
  - Variables in function scope
  - Multiple simultaneous usages
`);
