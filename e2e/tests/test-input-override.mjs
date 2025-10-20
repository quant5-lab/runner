#!/usr/bin/env node
/**
 * E2E Test: Input parameter overrides with DETERMINISTIC data validation
 *
 * Tests that inputOverrides parameter actually affects calculations by:
 * 1. Using MockProvider with predictable data (close = [1, 2, 3, 4, ...])
 * 2. Running same strategy with default and overridden input values
 * 3. Asserting outputs differ when input values differ
 * 4. Validating exact computed values match expected results
 */
import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';
import { MockProviderManager } from '../mocks/MockProvider.js';
import { readFile } from 'fs/promises';
import { strict as assert } from 'assert';
import { spawn } from 'child_process';

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test: Input Overrides with Deterministic Data');
console.log('═══════════════════════════════════════════════════════════\n');

/* Transpile Pine code to JavaScript */
async function transpilePineCode(pineCode) {
  return new Promise((resolve, reject) => {
    const timestamp = Date.now();
    const inputPath = `/tmp/input-${timestamp}.pine`;
    const outputPath = `/tmp/output-${timestamp}.json`;

    import('fs/promises').then(async (fs) => {
      await fs.writeFile(inputPath, pineCode, 'utf-8');

      const pythonProcess = spawn('python3', [
        'services/pine-parser/parser.py',
        inputPath,
        outputPath,
      ]);

      let stderr = '';
      pythonProcess.stderr.on('data', (data) => {
        stderr += data.toString();
      });

      pythonProcess.on('close', async (code) => {
        if (code !== 0) {
          reject(new Error(`Parser failed: ${stderr}`));
          return;
        }

        try {
          const astJson = await fs.readFile(outputPath, 'utf-8');
          const ast = JSON.parse(astJson);

          const escodegen = (await import('escodegen')).default;
          const jsCode = escodegen.generate(ast);

          resolve(jsCode);
        } catch (error) {
          reject(error);
        }
      });
    });
  });
}

/* Calculate expected SMA manually */
function calculateExpectedSMA(closes, period) {
  const result = [];
  for (let i = 0; i < closes.length; i++) {
    if (i < period - 1) {
      result.push(null);
    } else {
      const sum = closes.slice(i - period + 1, i + 1).reduce((a, b) => a + b, 0);
      result.push(sum / period);
    }
  }
  return result;
}

/* Run strategy with optional input overrides */
async function runStrategyWithOverrides(pineCode, inputOverrides = null) {
  const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 1 });
  const constructorOptions = inputOverrides ? { inputOverrides } : undefined;
  const pineTS = new PineTS(
    mockProvider,
    'TEST',
    'D',
    30,
    null,
    null,
    constructorOptions,
  );

  const jsCode = await transpilePineCode(pineCode);
  const { plotAdapterSource } = await import('../../src/adapters/PinePlotAdapter.js');

  const wrappedCode = `(context) => {
    const { close, open, high, low, volume } = context.data;
    const { plot: corePlot, color, na, nz } = context.core;
    const ta = context.ta;
    const math = context.math;
    const input = context.input;
    const syminfo = context.syminfo;
    
    ${plotAdapterSource}
    
    function indicator() {}
    function strategy() {}
    
    ${jsCode}
  }`;

  return await pineTS.run(wrappedCode);
}

/* Test: Input override changes output */
async function testInputOverride() {
  console.log('TEST 1: Input override produces different output\n');

  const pineCode = await readFile('e2e/fixtures/strategies/test-input-int.pine', 'utf-8');

  /* Run with default values */
  const resultDefault = await runStrategyWithOverrides(pineCode, null);
  const smaDefault = resultDefault.plots['SMA with named defval'].data.map((d) => d.value);

  /* Run with override: length1 = 10 instead of 14 */
  const resultOverride = await runStrategyWithOverrides(pineCode, {
    'Length 1 (named defval)': 10,
  });
  const smaOverride = resultOverride.plots['SMA with named defval'].data.map((d) => d.value);

  /* Outputs should differ */
  const nonNullDefault = smaDefault.filter((v) => v !== null && !isNaN(v)).length;
  const nonNullOverride = smaOverride.filter((v) => v !== null && !isNaN(v)).length;

  console.log(`  Default SMA(14): ${nonNullDefault} non-null values`);
  console.log(`  Override SMA(10): ${nonNullOverride} non-null values`);

  /* SMA(14) starts at bar 14 (17 values), SMA(10) starts at bar 10 (21 values) */
  assert.strictEqual(nonNullDefault, 17, 'Default SMA(14) should have 17 values');
  assert.strictEqual(nonNullOverride, 21, 'Override SMA(10) should have 21 values');

  /* Validate values match expected calculations */
  const closes = Array.from({ length: 30 }, (_, i) => i + 1);
  const expectedSMA14 = calculateExpectedSMA(closes, 14);
  const expectedSMA10 = calculateExpectedSMA(closes, 10);

  for (let i = 0; i < 30; i++) {
    if (expectedSMA14[i] !== null) {
      assert.ok(
        Math.abs(smaDefault[i] - expectedSMA14[i]) < 0.0001,
        `Default SMA14[${i}] should be ${expectedSMA14[i]}, got ${smaDefault[i]}`,
      );
    }

    if (expectedSMA10[i] !== null) {
      assert.ok(
        Math.abs(smaOverride[i] - expectedSMA10[i]) < 0.0001,
        `Override SMA10[${i}] should be ${expectedSMA10[i]}, got ${smaOverride[i]}`,
      );
    }
  }

  console.log('  ✅ Default values produce correct SMA(14)');
  console.log('  ✅ Override values produce correct SMA(10)');
  console.log('  ✅ Input overrides successfully change calculations\n');
}

/* Test: Multiple overrides */
async function testMultipleOverrides() {
  console.log('TEST 2: Multiple input overrides\n');

  const pineCode = await readFile('e2e/fixtures/strategies/test-input-float.pine', 'utf-8');

  /* Run with defaults: mult1=1.4, mult2=2.0 */
  const resultDefault = await runStrategyWithOverrides(pineCode, null);
  const sma14Default = resultDefault.plots['SMA (named defval)'].data.map((d) => d.value);
  const sma20Default = resultDefault.plots['SMA (defval first)'].data.map((d) => d.value);

  /* Run with overrides: mult1=2.0, mult2=1.5 */
  const resultOverride = await runStrategyWithOverrides(pineCode, {
    'Multiplier 1 (named defval)': 2.0,
    'Multiplier 2 (defval first)': 1.5,
  });
  const sma20Override = resultOverride.plots['SMA (named defval)'].data.map((d) => d.value);
  const sma15Override = resultOverride.plots['SMA (defval first)'].data.map((d) => d.value);

  const nonNullDefault14 = sma14Default.filter((v) => v !== null && !isNaN(v)).length;
  const nonNullDefault20 = sma20Default.filter((v) => v !== null && !isNaN(v)).length;
  const nonNullOverride20 = sma20Override.filter((v) => v !== null && !isNaN(v)).length;
  const nonNullOverride15 = sma15Override.filter((v) => v !== null && !isNaN(v)).length;

  console.log(`  Default: SMA(14)=${nonNullDefault14} values, SMA(20)=${nonNullDefault20} values`);
  console.log(
    `  Override: SMA(20)=${nonNullOverride20} values, SMA(15)=${nonNullOverride15} values`,
  );

  /* SMA(14)=17 values, SMA(20)=11 values, SMA(15)=16 values */
  assert.strictEqual(nonNullDefault14, 17, 'Default mult1*10=14 should give 17 values');
  assert.strictEqual(nonNullDefault20, 11, 'Default mult2*10=20 should give 11 values');
  assert.strictEqual(nonNullOverride20, 11, 'Override mult1*10=20 should give 11 values');
  assert.strictEqual(nonNullOverride15, 16, 'Override mult2*10=15 should give 16 values');

  console.log('  ✅ Multiple overrides successfully applied');
  console.log('  ✅ Each override produces correct period\n');
}

/* Run tests */
async function runTests() {
  try {
    await testInputOverride();
    await testMultipleOverrides();

    console.log('═══════════════════════════════════════════════════════════');
    console.log('✅ ALL INPUT OVERRIDE TESTS PASSED');
    console.log('═══════════════════════════════════════════════════════════');
    console.log('\nRegression protection: ✅ VALIDATED');
    console.log('  - Input overrides successfully change calculations');
    console.log('  - Default and override values produce expected results');
    console.log('  - Multiple overrides work correctly');
    console.log('  - No network dependencies (100% deterministic)');

    process.exit(0);
  } catch (error) {
    console.error('\n═══════════════════════════════════════════════════════════');
    console.error('❌ TEST FAILED');
    console.error('═══════════════════════════════════════════════════════════');
    console.error(error.message);
    console.error(error.stack);
    process.exit(1);
  }
}

runTests();
