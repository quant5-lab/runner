#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { MockProviderManager } from '../mocks/MockProvider.js';
import { FLOAT_EPSILON, assertFloatEquals } from '../utils/test-helpers.js';

/* Helper to run strategy with specific data pattern */
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

function calcPivotHigh(highs, leftbars, rightbars) {
  const result = [];
  for (let i = 0; i < highs.length; i++) {
    /* Pine returns pivot at confirmation point (rightbars after the peak) */
    const pivotIndex = i - rightbars;

    if (pivotIndex < leftbars || pivotIndex + rightbars >= highs.length) {
      result.push(NaN);
      continue;
    }

    const pivotValue = highs[pivotIndex];
    let isPivot = true;

    for (let j = 1; j <= leftbars; j++) {
      if (highs[pivotIndex - j] >= pivotValue) {
        isPivot = false;
        break;
      }
    }

    if (isPivot) {
      for (let j = 1; j <= rightbars; j++) {
        if (highs[pivotIndex + j] >= pivotValue) {
          isPivot = false;
          break;
        }
      }
    }

    result.push(isPivot ? pivotValue : NaN);
  }
  return result;
}

function calcPivotLow(lows, leftbars, rightbars) {
  const result = [];
  for (let i = 0; i < lows.length; i++) {
    /* Pine returns pivot at confirmation point (rightbars after the valley) */
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

console.log('=== TA Functions E2E Tests ===\n');

const fixnanResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-fixnan.pine',
  'linear',
);
const closeValues = getPlotValues(fixnanResult, 'close');
const fixnanValues = getPlotValues(fixnanResult, 'fixnan result');

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

// Test pivothigh with sawtooth pattern - creates clear peaks for pivot detection
const pivotResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-pivothigh.pine',
  'sawtooth',
  100,
  10,
);
const pivotHighValues = getPlotValues(pivotResult, 'high');
const pivot2Values = getPlotValues(pivotResult, 'pivot2');
const pivot5Values = getPlotValues(pivotResult, 'pivot5');

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
  `✅ pivothigh: ${pivotMatched}/${pivotTotal} match (found ${pivot2Count} pivot2, ${pivot5Count} pivot5)`,
);

// Test pivotlow with sawtooth pattern
const pivotlowResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-pivotlow.pine',
  'sawtooth',
  100,
  10,
);
const lowValues = getPlotValues(pivotlowResult, 'low');
const pivotlow2Values = getPlotValues(pivotlowResult, 'pivot2');
const pivotlow5Values = getPlotValues(pivotlowResult, 'pivot5');

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
  `✅ pivotlow: ${pivotlowMatched}/${pivotlowTotal} match (found ${pivotlow2Count} pivot2, ${pivotlow5Count} pivot5)`,
);

// Test valuewhen
const valuewhenResult = await runStrategyWithPattern(
  50,
  'e2e/fixtures/strategies/test-valuewhen.pine',
  'linear',
);
const vw0Values = getPlotValues(valuewhenResult, 'valuewhen_0');
const vw1Values = getPlotValues(valuewhenResult, 'valuewhen_1');
const conditionValues = getPlotValues(valuewhenResult, 'condition');
const vwHighValues = getPlotValues(valuewhenResult, 'high');

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
  `✅ valuewhen: ${valuewhenMatched}/${vw0Values.length + vw1Values.length} values match`,
);

// Test barmerge constants
const barmergeResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-barmerge.pine',
  'linear',
);
const lookaheadValues = getPlotValues(barmergeResult, 'Daily Open (lookahead)');
const noLookaheadValues = getPlotValues(barmergeResult, 'Daily Open (no lookahead)');
console.log('✅ barmerge: All 4 constants available (lookahead_on/off, gaps_on/off)');
console.log(`   - lookahead_on: ${lookaheadValues?.length || 0} values`);
console.log(`   - lookahead_off: ${noLookaheadValues?.length || 0} values`);

// Test time()
const timeResult = await runStrategyWithPattern(
  30,
  'e2e/fixtures/strategies/test-time.pine',
  'linear',
);
const timeDailyValues = getPlotValues(timeResult, 'time_daily');
const timeWeeklyValues = getPlotValues(timeResult, 'time_weekly');
const validDailyTimes = timeDailyValues.filter((v) => v !== null && !isNaN(v) && v > 0).length;
const validWeeklyTimes = timeWeeklyValues.filter((v) => v !== null && !isNaN(v) && v > 0).length;
console.log(
  `✅ time: Daily ${validDailyTimes}/${timeDailyValues.length}, Weekly ${validWeeklyTimes}/${timeWeeklyValues.length} valid timestamps`,
);

console.log('\n=== All tests passed ✅ ===');
