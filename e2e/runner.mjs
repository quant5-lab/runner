#!/usr/bin/env node
import { spawn } from 'child_process';
import { readdir } from 'fs/promises';
import { join, basename, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const TESTS_DIR = join(__dirname, 'tests');
const TIMEOUT_MS = 60000;

class TestRunner {
  constructor() {
    this.results = [];
    this.startTime = Date.now();
  }

  async discoverTests() {
    const files = await readdir(TESTS_DIR);
    return files
      .filter((f) => f.endsWith('.mjs') && !f.endsWith('.bak'))
      .sort()
      .map((f) => join(TESTS_DIR, f));
  }

  async runTest(testPath) {
    const testName = basename(testPath);
    const startTime = Date.now();

    return new Promise((resolve) => {
      const child = spawn('node', [testPath], {
        stdio: ['ignore', 'pipe', 'pipe'],
        timeout: TIMEOUT_MS,
      });

      let stdout = '';
      let stderr = '';

      child.stdout.on('data', (data) => {
        stdout += data.toString();
      });

      child.stderr.on('data', (data) => {
        stderr += data.toString();
      });

      const timer = setTimeout(() => {
        child.kill('SIGTERM');
      }, TIMEOUT_MS);

      child.on('close', (code) => {
        clearTimeout(timer);
        const duration = Date.now() - startTime;

        resolve({
          name: testName,
          path: testPath,
          passed: code === 0,
          exitCode: code,
          duration,
          stdout,
          stderr,
        });
      });

      child.on('error', (error) => {
        clearTimeout(timer);
        const duration = Date.now() - startTime;

        resolve({
          name: testName,
          path: testPath,
          passed: false,
          exitCode: -1,
          duration,
          stdout,
          stderr: error.message,
        });
      });
    });
  }

  async runAll() {
    console.log('═══════════════════════════════════════════════════════════');
    console.log('E2E Test Suite');
    console.log('═══════════════════════════════════════════════════════════\n');

    const tests = await this.discoverTests();
    console.log(`Discovered ${tests.length} tests\n`);

    for (const testPath of tests) {
      const testName = basename(testPath);
      console.log(`Running: ${testName}`);

      const result = await this.runTest(testPath);
      this.results.push(result);

      if (result.passed) {
        console.log(`✅ PASS (${result.duration}ms)\n`);
      } else {
        console.log(`❌ FAIL (${result.duration}ms)`);
        if (result.stderr) {
          console.log(`Error output:\n${result.stderr}\n`);
        }
      }
    }

    this.printSummary();
    return this.getFailureCount() === 0;
  }

  getFailureCount() {
    return this.results.filter((r) => !r.passed).length;
  }

  getPassCount() {
    return this.results.filter((r) => r.passed).length;
  }

  getTotalDuration() {
    return Date.now() - this.startTime;
  }

  printSummary() {
    const passed = this.getPassCount();
    const failed = this.getFailureCount();
    const total = this.results.length;
    const duration = this.getTotalDuration();

    console.log('═══════════════════════════════════════════════════════════');
    console.log('Test Summary');
    console.log('═══════════════════════════════════════════════════════════');
    console.log(`Total:    ${total}`);
    console.log(`Passed:   ${passed} (${((passed / total) * 100).toFixed(1)}%)`);
    console.log(`Failed:   ${failed} (${((failed / total) * 100).toFixed(1)}%)`);
    console.log(`Duration: ${(duration / 1000).toFixed(2)}s\n`);

    if (failed > 0) {
      console.log('Failed Tests:');
      this.results
        .filter((r) => !r.passed)
        .forEach((r) => {
          console.log(`  ❌ ${r.name} (exit code: ${r.exitCode})`);
        });
      console.log('');
    }

    if (failed === 0) {
      console.log('✅ ALL TESTS PASSED\n');
    } else {
      console.log('❌ SOME TESTS FAILED\n');
    }
  }
}

async function main() {
  const runner = new TestRunner();
  const success = await runner.runAll();
  process.exit(success ? 0 : 1);
}

main().catch((error) => {
  console.error('Fatal error in test runner:');
  console.error(error);
  process.exit(1);
});
