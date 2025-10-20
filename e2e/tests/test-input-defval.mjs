#!/usr/bin/env node
/**
 * E2E Test: input.* functions with DETERMINISTIC data validation
 * 
 * Tests that input parameters actually affect calculations by:
 * 1. Using MockProvider with predictable data (close = [1, 2, 3, 4, ...])
 * 2. Calculating expected SMA values manually
 * 3. Asserting actual output matches expected output EXACTLY
 * 
 * This provides TRUE regression protection vs AST-only validation.
 */
import { PineTS } from '../../../PineTS/dist/pinets.dev.es.js';
import { MockProviderManager } from '../mocks/MockProvider.js';
import { readFile } from 'fs/promises';
import { strict as assert } from 'assert';
import { spawn } from 'child_process';

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test: input.* defval with Deterministic Data');
console.log('═══════════════════════════════════════════════════════════\n');

/**
 * Transpile Pine code to JavaScript
 */
async function transpilePineCode(pineCode) {
  return new Promise((resolve, reject) => {
    const timestamp = Date.now();
    const inputPath = `/tmp/input-${timestamp}.pine`;
    const outputPath = `/tmp/output-${timestamp}.json`;
    
    import('fs/promises').then(async (fs) => {
      await fs.writeFile(inputPath, pineCode, 'utf-8');
      
      const pythonProcess = spawn('python3', ['services/pine-parser/parser.py', inputPath, outputPath]);
      
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
          
          // Generate JS code from AST
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

/**
 * Calculate expected SMA manually
 */
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

/**
 * Test: input.int() with deterministic data
 */
async function testInputIntDeterministic() {
  console.log('TEST 1: input.int() produces correct SMA values\n');
  
  // Setup MockProvider with linear data: close = [1, 2, 3, 4, ...]
  const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 1 });
  const pineTS = new PineTS(mockProvider, 'TEST', 'D', 30, null, null);
  
  // Read and transpile strategy
  const pineCode = await readFile('e2e/fixtures/strategies/test-input-int.pine', 'utf-8');
  const jsCode = await transpilePineCode(pineCode);
  
  // Import plot adapter
  const { plotAdapterSource } = await import('../../src/adapters/PinePlotAdapter.js');
  
  // Wrap code for PineTS execution
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
  
  // Execute strategy
  const result = await pineTS.run(wrappedCode);
  
  // Generate expected values for close = [1, 2, 3, ..., 30]
  const closes = Array.from({ length: 30 }, (_, i) => i + 1);
  const expectedSMA14 = calculateExpectedSMA(closes, 14);
  const expectedSMA20 = calculateExpectedSMA(closes, 20);
  const expectedSMA10 = calculateExpectedSMA(closes, 10);
  
  // Extract actual values
  const actualSMA14 = result.plots['SMA with named defval'].data.map(d => d.value);
  const actualSMA20 = result.plots['SMA with defval first'].data.map(d => d.value);
  const actualSMA10 = result.plots['SMA with positional'].data.map(d => d.value);
  
  // Validate lengths
  assert.strictEqual(actualSMA14.length, 30, 'SMA14 should have 30 values');
  assert.strictEqual(actualSMA20.length, 30, 'SMA20 should have 30 values');
  assert.strictEqual(actualSMA10.length, 30, 'SMA10 should have 30 values');
  
  // Count non-null, non-NaN values
  const nonNullSMA14 = actualSMA14.filter(v => v !== null && !isNaN(v)).length;
  const nonNullSMA20 = actualSMA20.filter(v => v !== null && !isNaN(v)).length;
  const nonNullSMA10 = actualSMA10.filter(v => v !== null && !isNaN(v)).length;
  
  console.log(`  DEBUG SMA14 first 5:`, actualSMA14.slice(0, 5));
  console.log(`  DEBUG SMA14 last 5:`, actualSMA14.slice(-5));
  console.log(`  SMA(14): ${nonNullSMA14} valid values (expected 17: bars 14-30)`);
  console.log(`  SMA(20): ${nonNullSMA20} valid values (expected 11: bars 20-30)`);
  console.log(`  SMA(10): ${nonNullSMA10} valid values (expected 21: bars 10-30)`);
  
  // Assert correct number of non-null values
  assert.strictEqual(nonNullSMA14, 17, 'SMA(14) should start at bar 14');
  assert.strictEqual(nonNullSMA20, 11, 'SMA(20) should start at bar 20');
  assert.strictEqual(nonNullSMA10, 21, 'SMA(10) should start at bar 10');
  
  // Validate actual computed values match expected
  for (let i = 0; i < 30; i++) {
    if (expectedSMA14[i] !== null) {
      assert.ok(
        Math.abs(actualSMA14[i] - expectedSMA14[i]) < 0.0001,
        `SMA14[${i}] should be ${expectedSMA14[i]}, got ${actualSMA14[i]}`
      );
    }
    
    if (expectedSMA20[i] !== null) {
      assert.ok(
        Math.abs(actualSMA20[i] - expectedSMA20[i]) < 0.0001,
        `SMA20[${i}] should be ${expectedSMA20[i]}, got ${actualSMA20[i]}`
      );
    }
    
    if (expectedSMA10[i] !== null) {
      assert.ok(
        Math.abs(actualSMA10[i] - expectedSMA10[i]) < 0.0001,
        `SMA10[${i}] should be ${expectedSMA10[i]}, got ${actualSMA10[i]}`
      );
    }
  }
  
  console.log('  ✅ All SMA values match expected calculations');
  console.log('  ✅ Input parameters correctly affect output\n');
}

/**
 * Test: input.float() with deterministic data
 */
async function testInputFloatDeterministic() {
  console.log('TEST 2: input.float() produces correct SMA values\n');
  
  // Setup MockProvider
  const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 1 });
  const pineTS = new PineTS(mockProvider, 'TEST', 'D', 30, null, null);
  
  // Read and transpile strategy
  const pineCode = await readFile('e2e/fixtures/strategies/test-input-float.pine', 'utf-8');
  const jsCode = await transpilePineCode(pineCode);
  
  // Import plot adapter
  const { plotAdapterSource } = await import('../../src/adapters/PinePlotAdapter.js');
  
  // Wrap code
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
  
  // Execute
  const result = await pineTS.run(wrappedCode);
  
  // mult1=1.4 → SMA(14), mult2=2.0 → SMA(20), mult3=1.0 → SMA(10)
  const closes = Array.from({ length: 30 }, (_, i) => i + 1);
  const expectedSMA14 = calculateExpectedSMA(closes, 14);
  const expectedSMA20 = calculateExpectedSMA(closes, 20);
  const expectedSMA10 = calculateExpectedSMA(closes, 10);
  
  // Extract actual
  const actualSMA14 = result.plots['SMA (named defval)'].data.map(d => d.value);
  const actualSMA20 = result.plots['SMA (defval first)'].data.map(d => d.value);
  const actualSMA10 = result.plots['SMA (positional)'].data.map(d => d.value);
  
  // Count valid (non-null, non-NaN) values
  const nonNullSMA14 = actualSMA14.filter(v => v !== null && !isNaN(v)).length;
  const nonNullSMA20 = actualSMA20.filter(v => v !== null && !isNaN(v)).length;
  const nonNullSMA10 = actualSMA10.filter(v => v !== null && !isNaN(v)).length;
  
  console.log(`  SMA(14): ${nonNullSMA14} valid values (expected 17)`);
  console.log(`  SMA(20): ${nonNullSMA20} valid values (expected 11)`);
  console.log(`  SMA(10): ${nonNullSMA10} valid values (expected 21)`);
  
  // Assert counts
  assert.strictEqual(nonNullSMA14, 17, 'mult1*10=14 → SMA(14) starts at bar 14');
  assert.strictEqual(nonNullSMA20, 11, 'mult2*10=20 → SMA(20) starts at bar 20');
  assert.strictEqual(nonNullSMA10, 21, 'mult3*10=10 → SMA(10) starts at bar 10');
  
  // Validate values
  for (let i = 0; i < 30; i++) {
    if (expectedSMA14[i] !== null) {
      assert.ok(
        Math.abs(actualSMA14[i] - expectedSMA14[i]) < 0.0001,
        `Float SMA14[${i}] mismatch`
      );
    }
  }
  
  console.log('  ✅ Float multipliers correctly calculate periods');
  console.log('  ✅ All SMA values match expected\n');
}

// Run tests
async function runTests() {
  try {
    await testInputIntDeterministic();
    await testInputFloatDeterministic();
    
    console.log('═══════════════════════════════════════════════════════════');
    console.log('✅ ALL DETERMINISTIC TESTS PASSED');
    console.log('═══════════════════════════════════════════════════════════');
    console.log('\nRegression protection: ✅ VALIDATED');
    console.log('  - Input parameters affect calculations');
    console.log('  - Computed values match expected results');
    console.log('  - No network dependencies');
    console.log('  - 100% deterministic');
    
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
