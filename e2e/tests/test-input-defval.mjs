#!/usr/bin/env node
import { createContainer } from '../../src/container.js';
import { createProviderChain, DEFAULTS } from '../../src/config.js';
import { readFile } from 'fs/promises';
import { strict as assert } from 'assert';

const container = createContainer(createProviderChain, DEFAULTS);
const runner = container.resolve('tradingAnalysisRunner');
const transpiler = container.resolve('pineScriptTranspiler');

console.log('═══════════════════════════════════════════════════════════');
console.log('E2E Test Suite: input.* defval Parameter Positioning');
console.log('═══════════════════════════════════════════════════════════\n');

async function executeStrategy(strategyPath, symbol, timeframe, bars) {
  const pineCode = await readFile(strategyPath, 'utf-8');
  const jsCode = await transpiler.transpile(pineCode);
  
  try {
    const result = await runner.runPineScriptStrategy(symbol, timeframe, bars, jsCode, strategyPath);
    return { success: true, result: result };
  } catch (error) {
    return { success: false, error: error.message };
  }
}

async function transpileStrategy(strategyPath) {
  const pineCode = await readFile(strategyPath, 'utf-8');
  const { spawn } = await import('child_process');
  const { promisify } = await import('util');
  const execPromise = promisify(spawn);
  
  const timestamp = Date.now();
  const inputPath = `/tmp/input-${timestamp}.pine`;
  const outputPath = `/tmp/output-${timestamp}.json`;
  
  await import('fs/promises').then(fs => fs.writeFile(inputPath, pineCode, 'utf-8'));
  
  return new Promise((resolve, reject) => {
    const pythonProcess = spawn('python3', ['services/pine-parser/parser.py', inputPath, outputPath]);
    
    let stderr = '';
    pythonProcess.stderr.on('data', (data) => {
      stderr += data.toString();
    });
    
    pythonProcess.on('close', async (code) => {
      if (code !== 0) {
        reject(new Error(`Python parser failed: ${stderr}`));
        return;
      }
      
      try {
        const astJson = await readFile(outputPath, 'utf-8');
        const ast = JSON.parse(astJson);
        resolve(ast);
      } catch (error) {
        reject(new Error(`Failed to read AST: ${error.message}`));
      }
    });
  });
}

function validateInputCallInAST(ast, inputFunction, expectedDefval) {
  const declarations = ast.body.filter(node => node.type === 'VariableDeclaration');
  
  for (const decl of declarations) {
    for (const declarator of decl.declarations) {
      if (declarator.init && 
          declarator.init.type === 'CallExpression' &&
          declarator.init.callee.type === 'MemberExpression' &&
          declarator.init.callee.object.name === 'input' &&
          declarator.init.callee.property.name === inputFunction) {
        
        const args = declarator.init.arguments;
        
        assert.ok(args.length > 0, `${inputFunction}() should have at least one argument`);
        
        const firstArg = args[0];
        
        if (firstArg.type === 'Literal') {
          assert.strictEqual(
            firstArg.value,
            expectedDefval,
            `${inputFunction}() first arg should be defval=${expectedDefval}`
          );
        } else if (firstArg.type === 'Identifier') {
          assert.strictEqual(
            firstArg.name,
            expectedDefval,
            `${inputFunction}() first arg should be ${expectedDefval}`
          );
        }
        
        if (args.length > 1 && args[1].type === 'ObjectExpression') {
          const hasDefvalInOptions = args[1].properties.some(
            prop => prop.key.name === 'defval'
          );
          assert.ok(
            !hasDefvalInOptions,
            `${inputFunction}() should not have 'defval' in options object`
          );
        }
        
        return true;
      }
    }
  }
  
  throw new Error(`Could not find ${inputFunction}() call in AST`);
}

async function testInputInt() {
  console.log('TEST: input.int() with defval parameter positioning');
  
  const result = await executeStrategy('e2e/fixtures/strategies/test-input-int.pine', 'CHMF', 'Monthly', 20);
  
  if (!result.success) {
    console.log('  ❌ Strategy execution failed');
    console.log('  ERROR:', result.error);
  }
  
  assert.ok(result.success, 'Strategy execution should succeed');
  assert.ok(result.result.plots, 'Should have plots in result');
  
  const ast = await transpileStrategy('e2e/fixtures/strategies/test-input-int.pine');
  
  validateInputCallInAST(ast, 'int', 14);
  
  console.log('  ✅ input.int(title="X", defval=14) transpiled correctly');
  console.log('  ✅ Strategy executed successfully');
}

async function testInputFloat() {
  console.log('\nTEST: input.float() with defval parameter positioning');
  
  const result = await executeStrategy('e2e/fixtures/strategies/test-input-float.pine', 'CHMF', 'Monthly', 20);
  
  if (!result.success) {
    console.log('  ❌ Strategy execution failed');
    console.log('  ERROR:', result.error);
  }
  
  assert.ok(result.success, 'Strategy execution should succeed');
  assert.ok(result.result.plots, 'Should have plots in result');
  
  const ast = await transpileStrategy('e2e/fixtures/strategies/test-input-float.pine');
  
  validateInputCallInAST(ast, 'float', 1.4);
  
  console.log('  ✅ input.float(title="Y", defval=1.4) transpiled correctly');
  console.log('  ✅ Strategy executed successfully');
}

async function testInputSourceRegression() {
  console.log('\nTEST: input.source() regression test');
  
  const result = await executeStrategy('strategies/rolling-cagr.pine', 'SBER', 'D', 24);
  
  if (!result.success) {
    console.log('  ❌ Strategy execution failed');
    console.log('  ERROR:', result.error);
  }
  
  assert.ok(result.success, 'Strategy execution should succeed');
  assert.ok(result.result.plots, 'Should have plots in result');
  
  const cagrPlot = result.result.plots['CAGR A'];
  assert.ok(cagrPlot, 'CAGR A plot should exist');
  
  const nonNullValues = cagrPlot.data.filter(d => d.value !== null);
  assert.ok(
    nonNullValues.length > 0,
    'Should have at least some calculated CAGR values'
  );
  
  console.log(`  ✅ input.source() still works (${nonNullValues.length} CAGR values calculated)`);
  console.log('  ✅ No regression detected');
}

async function runTests() {
  try {
    await testInputInt();
    await testInputFloat();
    
    console.log('\n═══════════════════════════════════════════════════════════');
    console.log('✅ ALL TESTS PASSED');
    console.log('═══════════════════════════════════════════════════════════');
    
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
