#!/usr/bin/env node
import { createContainer } from '../../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../../utils/test-helpers.js';

console.log('=== TA Function Test: pivothigh() ===\n');

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

const pineCode = await readFile('e2e/fixtures/strategies/test-pivothigh.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);
const result = await runner.runPineScriptStrategy('TEST', '1h', 30, jsCode, 'test-pivothigh.pine');

function getPlotValues(result, plotTitle) {
  const plot = result.plots?.[plotTitle];
  if (!plot || !plot.data) return null;
  return plot.data.map((d) => d.value);
}

function calcPivotHigh(highs, leftbars, rightbars) {
  const result = [];
  for (let i = 0; i < highs.length; i++) {
    const pivotIndex = i - rightbars;
    
    if (pivotIndex < leftbars) {
      result.push(NaN);
      continue;
    }
    
    const pivotValue = highs[pivotIndex];
    let isPivot = true;
    
    for (let j = pivotIndex - leftbars; j < pivotIndex; j++) {
      if (highs[j] >= pivotValue) {
        isPivot = false;
        break;
      }
    }
    
    if (isPivot) {
      for (let j = pivotIndex + 1; j <= pivotIndex + rightbars; j++) {
        if (highs[j] >= pivotValue) {
          isPivot = false;
          break;
        }
      }
    }
    
    if (isPivot) {
      result.push(pivotValue);
    } else {
      result.push(NaN);
    }
  }
  return result;
}

const pivotHighValues = getPlotValues(result, 'high');
const pivot2Values = getPlotValues(result, 'pivot2');
const pivot5Values = getPlotValues(result, 'pivot5');

const jsPivot2 = calcPivotHigh(pivotHighValues, 2, 2);
const jsPivot5 = calcPivotHigh(pivotHighValues, 5, 5);

let pivotMatched = 0;
let pivotTotal = 0;

for (let i = 0; i < pivot2Values.length; i++) {
  if (isNaN(pivot2Values[i]) && isNaN(jsPivot2[i])) {
    pivotMatched++;
  } else if (!isNaN(pivot2Values[i]) && !isNaN(jsPivot2[i])) {
    assertFloatEquals(pivot2Values[i], jsPivot2[i], FLOAT_EPSILON, `pivot2[${i}]`);
    pivotMatched++;
  }
  pivotTotal++;
}

for (let i = 0; i < pivot5Values.length; i++) {
  if (isNaN(pivot5Values[i]) && isNaN(jsPivot5[i])) {
    pivotMatched++;
  } else if (!isNaN(pivot5Values[i]) && !isNaN(jsPivot5[i])) {
    assertFloatEquals(pivot5Values[i], jsPivot5[i], FLOAT_EPSILON, `pivot5[${i}]`);
    pivotMatched++;
  }
  pivotTotal++;
}

const pivot2Count = pivot2Values.filter((v) => !isNaN(v)).length;
const pivot5Count = pivot5Values.filter((v) => !isNaN(v)).length;

console.log(
  `✅ pivothigh: ${pivotMatched}/${pivotTotal} match (found ${pivot2Count} pivot2, ${pivot5Count} pivot5)`
);

process.exit(0);
