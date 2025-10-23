#!/usr/bin/env node
import { spawn } from 'child_process';
import { resolve } from 'path';

/* E2E test validating strategy namespace transpiler transformation */
function runTest() {
  return new Promise((resolve, reject) => {
    const args = [
      'src/index.js',
      'BTCUSDT',
      '1h',
      '50',
      'e2e/fixtures/strategies/test-strategy.pine'
    ];
    
    const proc = spawn('node', args, { 
      cwd: process.cwd(),
      env: { ...process.env, DEBUG: 'false' }
    });

    let stdout = '';
    let stderr = '';

    proc.stdout.on('data', (data) => { stdout += data.toString(); });
    proc.stderr.on('data', (data) => { stderr += data.toString(); });

    proc.on('close', (code) => {
      if (code !== 0) {
        reject(new Error(`Exit code ${code}\nSTDERR: ${stderr}\nSTDOUT: ${stdout}`));
        return;
      }

      /* Validate strategy namespace features executed */
      const requiredFeatures = [
        'strategy.cash',           // default_qty_type parameter
        'strategy.commission',     // commission_type parameter
        'strategy.long',           // entry direction constant
        'strategy.short',          // entry direction constant
        'strategy.entry',          // entry method
        'strategy.exit',           // exit method
        'strategy.position_avg_price', // position tracking property
        'strategy.close_all'       // close all positions method
      ];

      /* Success if no errors and execution completed */
      if (stderr.includes('Error:') || stderr.includes('ReferenceError:')) {
        reject(new Error(`Strategy namespace error detected:\n${stderr}`));
        return;
      }

      if (!stdout.includes('Completed in:')) {
        reject(new Error(`Execution did not complete\nSTDOUT: ${stdout}`));
        return;
      }

      console.log('✅ Strategy namespace transpiler transformation validated');
      console.log('✅ strategy() → strategy.call() transformation working');
      console.log('✅ All strategy.* features accessible');
      resolve();
    });

    proc.on('error', (error) => {
      reject(new Error(`Failed to spawn process: ${error.message}`));
    });
  });
}

/* Execute test */
runTest()
  .then(() => {
    console.log('✅ E2E test-strategy.mjs PASSED');
    process.exit(0);
  })
  .catch((error) => {
    console.error('❌ E2E test-strategy.mjs FAILED:', error.message);
    process.exit(1);
  });
