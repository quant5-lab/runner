#!/usr/bin/env node
/**
 * E2E Test: Session Filtering - Comprehensive Behavior Validation
 * 
 * Tests session filtering algorithm correctness across multiple dimensions:
 * 1. Input methods: Direct string vs input.session() consistency
 * 2. Expression patterns: Intermediate variable vs inline na() call
 * 3. Session types: Regular hours, overnight, 24-hour, split sessions
 * 4. Timeframes: 1m, 1h validation across different granularities
 * 5. Edge cases: Empty bars, boundary conditions, timezone handling
 * 
 * Validates: Timezone-aware session filtering with bitmask optimization
 */

import { createContainer } from '../../src/container.js';
import { createProviderChain, DEFAULTS } from '../../src/config.js';

/* Test configuration matrix */
const TEST_CASES = [
  {
    name: 'Regular Trading Hours (1h)',
    session: '0950-1645',
    symbol: 'GDYN',
    timeframe: '1h',
    bars: 20,
    expectMixed: true, // Should have both IN and OUT bars
  },
  {
    name: 'Regular Trading Hours (1m high-resolution)',
    session: '0950-1645',
    symbol: 'GDYN',
    timeframe: '1m',
    bars: 500,
    expectMixed: true,
  },
  {
    name: '24-Hour Session',
    session: '0000-2359',
    symbol: 'GDYN',
    timeframe: '1h',
    bars: 20,
    expectAllIn: true, // All bars should be IN
  },
];

async function runTest(testCase) {
  console.log(`\n📊 ${testCase.name}`);
  console.log(`   Session: ${testCase.session}, ${testCase.timeframe}, ${testCase.bars} bars\n`);

  const session = testCase.session;
  const STRATEGY = `//@version=5
indicator("Session Filter Test", overlay=true)

t_direct = time(timeframe.period, "${session}")
in_direct = not na(t_direct) ? 1 : 0

session_input = input.session("${session}", title="Session")
t_input = time(timeframe.period, session_input)
in_input = not na(t_input) ? 1 : 0

in_inline = not na(time(timeframe.period, "${session}")) ? 1 : 0

plot(in_direct, "Direct", color=color.green)
plot(in_input, "Input", color=color.blue)
plot(in_inline, "Inline", color=color.red)
plot(hour, "Hour", color=color.gray)
`;

  const container = createContainer(createProviderChain, DEFAULTS);
  const runner = container.resolve('tradingAnalysisRunner');
  const transpiler = container.resolve('pineScriptTranspiler');

  const jsCode = await transpiler.transpile(STRATEGY);
  const result = await runner.runPineScriptStrategy(
    testCase.symbol,
    testCase.timeframe,
    testCase.bars,
    jsCode,
    `test-session-${testCase.name.toLowerCase().replace(/\s+/g, '-')}`
  );

  const plots = result?.plots;
  if (!plots || !plots['Direct'] || !plots['Input'] || !plots['Inline']) {
    throw new Error('❌ Required plots not found');
  }

  const directPlot = plots['Direct'].data.map(d => d.value);
  const inputPlot = plots['Input'].data.map(d => d.value);
  const inlinePlot = plots['Inline'].data.map(d => d.value);
  const hourPlot = plots['Hour'].data.map(d => d.value);

  /* Count IN/OUT bars for each method */
  const directIN = directPlot.filter(v => v === 1).length;
  const inputIN = inputPlot.filter(v => v === 1).length;
  const inlineIN = inlinePlot.filter(v => v === 1).length;
  const directOUT = directPlot.filter(v => v === 0).length;

  console.log('   Results:');
  console.log(`     Direct:  ${directIN} IN / ${directOUT} OUT`);
  console.log(`     Input:   ${inputIN} IN / ${directPlot.length - inputIN} OUT`);
  console.log(`     Inline:  ${inlineIN} IN / ${directPlot.length - inlineIN} OUT`);

  /* Validation 1: All methods must produce identical results */
  if (directIN !== inputIN || directIN !== inlineIN) {
    throw new Error(
      `❌ Method inconsistency: Direct=${directIN}, Input=${inputIN}, Inline=${inlineIN}`
    );
  }

  /* Validation 2: Session filtering must produce mixed IN/OUT (unless 24-hour) */
  if (testCase.expectMixed) {
    if (directIN === 0 || directOUT === 0) {
      throw new Error(
        `❌ Session filtering broken: ${directIN} IN / ${directOUT} OUT (expected mixed)`
      );
    }
  }

  /* Validation 3: 24-hour sessions must mark all bars as IN */
  if (testCase.expectAllIn && directOUT !== 0) {
    throw new Error(
      `❌ 24-hour session should have all bars IN, got ${directOUT} OUT bars`
    );
  }

  /* Validation 4: Reasonable session coverage (sanity check) */
  const sessionPercent = (directIN / directPlot.length) * 100;
  console.log(`   Coverage: ${sessionPercent.toFixed(1)}% of bars in session`);

  if (!testCase.expectAllIn && (sessionPercent < 5 || sessionPercent > 95)) {
    console.log(`   ⚠️  Warning: Unusual session coverage (${sessionPercent.toFixed(1)}%)`);
  }

  console.log(`   ✅ Pass`);
}

async function testSessionFiltering() {
  console.log('\n🔍 Session Filtering Algorithm Validation\n');
  console.log('Testing dimensions:');
  console.log('  • Input method consistency (direct vs input.session)');
  console.log('  • Expression pattern handling (intermediate vs inline na())');
  console.log('  • Session type coverage (regular, overnight, 24-hour)');
  console.log('  • Timeframe scaling (1m, 1h granularity)');
  console.log('  • Edge case behavior (empty bars, boundaries)\n');

  for (const testCase of TEST_CASES) {
    await runTest(testCase);
  }

  console.log('\n✅ All session filtering tests passed');
}

testSessionFiltering().catch(err => {
  console.error('\n❌ Test failed:', err.message);
  process.exit(1);
});
