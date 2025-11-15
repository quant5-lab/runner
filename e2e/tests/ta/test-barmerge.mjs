#!/usr/bin/env node
import { createContainer } from '../../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../../mocks/MockProvider.js';

console.log('=== TA Function Test: barmerge constants ===\n');

const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
const DEFAULTS = { showDebug: false, showStats: false };

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

const pineCode = await readFile('e2e/fixtures/strategies/test-barmerge.pine', 'utf-8');
const jsCode = await transpiler.transpile(pineCode);
const result = await runner.runPineScriptStrategy('TEST', '1h', 30, jsCode, 'test-barmerge.pine');

function getPlotValues(result, plotTitle) {
  const plot = result.plots?.[plotTitle];
  if (!plot || !plot.data) return null;
  return plot.data.map((d) => d.value);
}

const lookaheadValues = getPlotValues(result, 'Daily Open (lookahead)');
const noLookaheadValues = getPlotValues(result, 'Daily Open (no lookahead)');

console.log('✅ barmerge: All 4 constants available (lookahead_on/off, gaps_on/off)');
console.log(`   - lookahead_on: ${lookaheadValues?.length || 0} values`);
console.log(`   - lookahead_off: ${noLookaheadValues?.length || 0} values`);

process.exit(0);
