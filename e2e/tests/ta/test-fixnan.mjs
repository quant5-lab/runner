#!/usr/bin/env node
import { createContainer } from '../../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../../utils/test-helpers.js';

console.log('=== TA Function Test: fixnan() ===\n');

const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

const pineCode = await readFile('e2e/fixtures/strategies/test-fixnan.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);
const result = await runner.runPineScriptStrategy('TEST', '1h', 30, jsCode, 'test-fixnan.pine');

function getPlotValues(result, plotTitle) {
  const plot = result.plots?.[plotTitle];
  if (!plot || !plot.data) return null;
  return plot.data.map((d) => d.value);
}

function calcFixnan(source) {
  let lastValid = null;
  const result = [];
  for (let i = 0; i < source.length; i++) {
    if (source[i] !== null && !isNaN(source[i])) {
      lastValid = source[i];
    }
    result.push(lastValid);
  }
  return result;
}

const closeValues = getPlotValues(result, 'close');
const fixnanValues = getPlotValues(result, 'fixnan result');

if (!fixnanValues) {
  console.error('ERROR: fixnan result plot not found');
  process.exit(1);
}

const jsFixnan = calcFixnan(closeValues);

let fixnanMatched = 0;
for (let i = 0; i < fixnanValues.length; i++) {
  assertFloatEquals(fixnanValues[i], jsFixnan[i], FLOAT_EPSILON, `fixnan[${i}]`);
  fixnanMatched++;
}

console.log(`✅ fixnan: ${fixnanMatched}/${fixnanValues.length} values match`);
process.exit(0);
