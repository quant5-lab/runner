#!/usr/bin/env node
import { createContainer } from '../../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../../utils/test-helpers.js';

console.log('=== TA Function Test: valuewhen() ===\n');

const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

const pineCode = await readFile('e2e/fixtures/strategies/test-valuewhen.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);
const result = await runner.runPineScriptStrategy('TEST', '1h', 50, jsCode, 'test-valuewhen.pine');

function getPlotValues(result, plotTitle) {
  const plot = result.plots?.[plotTitle];
  if (!plot || !plot.data) return null;
  return plot.data.map((d) => d.value);
}

function calcValueWhen(conditions, source, occurrence) {
  const result = [];
  for (let i = 0; i < conditions.length; i++) {
    const trueIndices = [];
    for (let j = i; j >= 0; j--) {
      if (conditions[j] > 0) {
        trueIndices.push(j);
        if (trueIndices.length > occurrence) break;
      }
    }

    if (trueIndices.length > occurrence) {
      result.push(source[trueIndices[occurrence]]);
    } else {
      result.push(NaN);
    }
  }
  return result;
}

const vw0Values = getPlotValues(result, 'valuewhen_0');
const vw1Values = getPlotValues(result, 'valuewhen_1');
const conditionValues = getPlotValues(result, 'condition');
const vwHighValues = getPlotValues(result, 'high');

const jsVw0 = calcValueWhen(conditionValues, vwHighValues, 0);
const jsVw1 = calcValueWhen(conditionValues, vwHighValues, 1);

let valuewhenMatched = 0;

for (let i = 0; i < vw0Values.length; i++) {
  if (isNaN(vw0Values[i]) && isNaN(jsVw0[i])) {
    valuewhenMatched++;
  } else if (!isNaN(vw0Values[i]) && !isNaN(jsVw0[i])) {
    assertFloatEquals(vw0Values[i], jsVw0[i], FLOAT_EPSILON, `vw0[${i}]`);
    valuewhenMatched++;
  }
}

for (let i = 0; i < vw1Values.length; i++) {
  if (isNaN(vw1Values[i]) && isNaN(jsVw1[i])) {
    valuewhenMatched++;
  } else if (!isNaN(vw1Values[i]) && !isNaN(jsVw1[i])) {
    assertFloatEquals(vw1Values[i], jsVw1[i], FLOAT_EPSILON, `vw1[${i}]`);
    valuewhenMatched++;
  }
}

console.log(
  `✅ valuewhen: ${valuewhenMatched}/${vw0Values.length + vw1Values.length} values match`
);

process.exit(0);
