#!/usr/bin/env node
import { createContainer } from '../../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../../utils/test-helpers.js';

console.log('=== TA Function Test: pivotlow() ===\n');

const mockProvider = new MockProviderManager({ 
  dataPattern: 'sawtooth', 
  basePrice: 100, 
  amplitude: 10 
});
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

const pineCode = await readFile('e2e/fixtures/strategies/test-pivotlow.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);
const result = await runner.runPineScriptStrategy('TEST', '1h', 30, jsCode, 'test-pivotlow.pine');

function getPlotValues(result, plotTitle) {
  const plot = result.plots?.[plotTitle];
  if (!plot || !plot.data) return null;
  return plot.data.map((d) => d.value);
}

function calcPivotLow(lows, leftbars, rightbars) {
  const result = [];
  for (let i = 0; i < lows.length; i++) {
    const pivotIndex = i - rightbars;

    if (pivotIndex < leftbars || pivotIndex + rightbars >= lows.length) {
      result.push(NaN);
      continue;
    }

    const pivotValue = lows[pivotIndex];
    let isPivot = true;

    for (let j = 1; j <= leftbars; j++) {
      if (lows[pivotIndex - j] <= pivotValue) {
        isPivot = false;
        break;
      }
    }

    if (isPivot) {
      for (let j = 1; j <= rightbars; j++) {
        if (lows[pivotIndex + j] <= pivotValue) {
          isPivot = false;
          break;
        }
      }
    }

    result.push(isPivot ? pivotValue : NaN);
  }
  return result;
}

const lowValues = getPlotValues(result, 'low');
const pivotlow2Values = getPlotValues(result, 'pivot2');
const pivotlow5Values = getPlotValues(result, 'pivot5');

const jsPivotLow2 = calcPivotLow(lowValues, 2, 2);
const jsPivotLow5 = calcPivotLow(lowValues, 5, 5);

let pivotlowMatched = 0;
let pivotlowTotal = 0;

for (let i = 0; i < pivotlow2Values.length; i++) {
  if (isNaN(pivotlow2Values[i]) && isNaN(jsPivotLow2[i])) {
    pivotlowMatched++;
  } else if (!isNaN(pivotlow2Values[i]) && !isNaN(jsPivotLow2[i])) {
    assertFloatEquals(pivotlow2Values[i], jsPivotLow2[i], FLOAT_EPSILON, `pivotlow2[${i}]`);
    pivotlowMatched++;
  }
  pivotlowTotal++;
}

for (let i = 0; i < pivotlow5Values.length; i++) {
  if (isNaN(pivotlow5Values[i]) && isNaN(jsPivotLow5[i])) {
    pivotlowMatched++;
  } else if (!isNaN(pivotlow5Values[i]) && !isNaN(jsPivotLow5[i])) {
    assertFloatEquals(pivotlow5Values[i], jsPivotLow5[i], FLOAT_EPSILON, `pivotlow5[${i}]`);
    pivotlowMatched++;
  }
  pivotlowTotal++;
}

const pivotlow2Count = pivotlow2Values.filter((v) => !isNaN(v)).length;
const pivotlow5Count = pivotlow5Values.filter((v) => !isNaN(v)).length;

console.log(
  `✅ pivotlow: ${pivotlowMatched}/${pivotlowTotal} match (found ${pivotlow2Count} pivot2, ${pivotlow5Count} pivot5)`
);

process.exit(0);
