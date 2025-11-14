import { describe, it, expect, beforeAll } from 'vitest';
import { createContainer } from '../../src/container.js';
import { DEFAULTS } from '../../src/config.js';
import { MockProviderManager } from '../../e2e/mocks/MockProvider.js';

/**
 * COMPREHENSIVE TEST COVERAGE: Array Built-in Variables Fix
 * 
 * Tests Python parser fix for wrapping array built-in variables with [0] access.
 * Ensures PineScript built-ins (bar_index, close, open, etc.) are accessed as scalars.
 */
describe('Array Built-in Variables: Parser Fix Validation', () => {
  let container;
  let runner;
  let transpiler;
  let mockProvider;

  beforeAll(() => {
    mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
    container = createContainer(() => [{ name: 'MockProvider', instance: mockProvider }], DEFAULTS);
    runner = container.resolve('tradingAnalysisRunner');
    transpiler = container.resolve('pineScriptTranspiler');
  });

  describe('bar_index access patterns', () => {
    it('should wrap bar_index in simple comparison', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index > 5
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] > 5');
      expect(jsCode).not.toMatch(/bar_index\s*>/);
    });

    it('should wrap bar_index in complex boolean expression - transpilation', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index > 5 and bar_index < 15 and bar_index != 10
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] > 5');
      expect(jsCode).toContain('bar_index[0] < 15');
      expect(jsCode).toContain('bar_index[0] !== 10');
    });

    it('should wrap bar_index in complex boolean expression - execution', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index > 5 and bar_index < 15 and bar_index != 10
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      const result = await runner.runPineScriptStrategy('TEST', '1h', 20, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      // Bars 6-14 except 10 should be 1
      expect(values[6]).toBe(1);
      expect(values[10]).toBe(0); // bar_index = 10, excluded
      expect(values[14]).toBe(1);
      expect(values[15]).toBe(0); // bar_index >= 15
    });

    it('should wrap bar_index in arithmetic operations', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index + 10
plot(result, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] + 10');
      
      const result = await runner.runPineScriptStrategy('TEST', '1h', 5, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      expect(values[0]).toBe(10); // bar_index[0]=0 + 10
      expect(values[4]).toBe(14); // bar_index[0]=4 + 10
    });

    it('should wrap bar_index in modulo operations', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index % 5 == 0
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] % 5');
      
      const result = await runner.runPineScriptStrategy('TEST', '1h', 15, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      expect(values[0]).toBe(1);  // 0 % 5 = 0
      expect(values[5]).toBe(1);  // 5 % 5 = 0
      expect(values[10]).toBe(1); // 10 % 5 = 0
      expect(values[3]).toBe(0);  // 3 % 5 = 3
    });

    it('should wrap bar_index in ternary expressions', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index < 10 ? bar_index * 2 : bar_index / 2
plot(result, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] < 10');
      expect(jsCode).toContain('bar_index[0] * 2');
      expect(jsCode).toContain('bar_index[0] / 2');
    });
  });

  describe('OHLCV built-in variables', () => {
    it('should wrap close in comparisons', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = close > 105
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('close[0] > 105');
    });

    it('should wrap open, high, low in expressions', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
bullish = close > open
range = high - low
plot(bullish ? 1 : 0, "bullish")
plot(range, "range")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('close[0] > open[0]');
      expect(jsCode).toContain('high[0] - low[0]');
    });

    it('should wrap volume in calculations', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
high_volume = volume > 2000
plot(high_volume ? 1 : 0, "high_volume")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('volume[0] > 2000');
    });

    it('should wrap time in comparisons', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = time > 5
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('time[0] > 5');
    });
  });

  describe('Derived built-in variables', () => {
    it('should wrap hl2 (high+low)/2', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = hl2 > 100
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('hl2[0] > 100');
    });

    it('should wrap hlc3 (high+low+close)/3', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = hlc3 > 100
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('hlc3[0] > 100');
    });

    it('should wrap ohlc4 (open+high+low+close)/4', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = ohlc4 > 100
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('ohlc4[0] > 100');
    });
  });

  describe('Mixed built-in and user variables', () => {
    it('should wrap only built-ins in transpiled code', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
my_index = bar_index + 1
result = my_index > 5 and bar_index > 4
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      // bar_index should be wrapped
      expect(jsCode).toContain('bar_index[0] + 1');
      expect(jsCode).toContain('bar_index[0] > 4');
      
      // my_index should NOT be wrapped (it's a user variable)
      expect(jsCode).toMatch(/my_index\s*>\s*5/);
      expect(jsCode).not.toContain('my_index[0]');
    });

    it('should execute correctly with mixed variables', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
my_index = bar_index + 1
result = my_index > 5 and bar_index > 4
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      const result = await runner.runPineScriptStrategy('TEST', '1h', 10, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      // my_index > 5 means bar_index > 4, AND bar_index > 4
      // Both conditions true when bar_index >= 5
      expect(values[5]).toBe(1);
      expect(values[9]).toBe(1);
      expect(values[4]).toBe(0);
    });

    it('should transpile built-ins in function parameters', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
calc(x) => x > 5 ? x * 2 : x
result = calc(bar_index)
plot(result, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      // bar_index passed to function should be wrapped
      expect(jsCode).toContain('calc(bar_index[0])');
    });

    it('should execute functions with built-in parameters', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
calc(x) => x > 5 ? x * 2 : x
result = calc(bar_index)
plot(result, "result")
`;
      const jsCode = await transpiler.transpile(code);
      const result = await runner.runPineScriptStrategy('TEST', '1h', 10, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      expect(values[3]).toBe(3);  // 3 <= 5, return 3
      expect(values[6]).toBe(12); // 6 > 5, return 6*2=12
      expect(values[9]).toBe(18); // 9 > 5, return 9*2=18
    });
  });

  describe('Edge cases', () => {
    it('should handle bar_index at bar 0', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index == 0
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      const result = await runner.runPineScriptStrategy('TEST', '1h', 5, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      expect(values[0]).toBe(1); // First bar
      expect(values[1]).toBe(0);
      expect(values[4]).toBe(0);
    });

    it('should handle multiple occurrences in same expression', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index * bar_index + bar_index
plot(result, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] * bar_index[0] + bar_index[0]');
      
      const result = await runner.runPineScriptStrategy('TEST', '1h', 5, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      expect(values[0]).toBe(0);  // 0*0+0 = 0
      expect(values[2]).toBe(6);  // 2*2+2 = 6
      expect(values[3]).toBe(12); // 3*3+3 = 12
    });

    it('should handle nested ternary with built-ins', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index < 5 ? 1 : bar_index < 10 ? 2 : 3
plot(result, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] < 5');
      expect(jsCode).toContain('bar_index[0] < 10');
      
      const result = await runner.runPineScriptStrategy('TEST', '1h', 15, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      expect(values[2]).toBe(1);  // bar < 5
      expect(values[7]).toBe(2);  // 5 <= bar < 10
      expect(values[12]).toBe(3); // bar >= 10
    });

    it('should handle last_bar_index', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
result = bar_index == last_bar_index
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0]');
      expect(jsCode).toContain('last_bar_index[0]');
    });
  });

  describe('Reassignment operator with built-ins', () => {
    it('should handle := with bar_index condition', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
value := bar_index > 5 ? value[1] : close * 0.01
plot(value * 1000, "value")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] > 5');
      expect(jsCode).toContain('close[0] * 0.01');
      
      const result = await runner.runPineScriptStrategy('TEST', '1h', 15, jsCode, 'test.pine');
      const values = result.plots.value.data.map(d => d.value / 1000);
      
      // First 6 bars should recalculate
      for (let i = 0; i < 6; i++) {
        expect(values[i]).toBeGreaterThan(0);
      }
      
      // After bar 6, should preserve
      const preserved = values[6];
      for (let i = 7; i < 15; i++) {
        expect(Math.abs(values[i] - preserved)).toBeLessThan(0.0001);
      }
    });
  });

  describe('Real-world patterns', () => {
    it('should handle bar counting pattern', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
bars_back = 10
target_bar = bar_index - bars_back
result = target_bar >= 0
plot(result ? 1 : 0, "result")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] - bars_back');
      
      const result = await runner.runPineScriptStrategy('TEST', '1h', 15, jsCode, 'test.pine');
      const values = result.plots.result.data.map(d => d.value);
      
      expect(values[0]).toBe(0);  // bar 0 - 10 < 0
      expect(values[9]).toBe(0);  // bar 9 - 10 < 0
      expect(values[10]).toBe(1); // bar 10 - 10 >= 0
      expect(values[14]).toBe(1); // bar 14 - 10 >= 0
    });

    it('should handle session filtering pattern', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
in_session = bar_index >= 5 and bar_index < 15
value = in_session ? close : 0
plot(value, "value")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('bar_index[0] >= 5');
      expect(jsCode).toContain('bar_index[0] < 15');
      expect(jsCode).toContain('close[0]');
    });

    it('should handle price crossover pattern', async () => {
      const code = `//@version=5
indicator("Test", overlay=true)
crossover = close > open and close[1] <= open[1]
plot(crossover ? 1 : 0, "crossover")
`;
      const jsCode = await transpiler.transpile(code);
      
      expect(jsCode).toContain('close[0] > open[0]');
      expect(jsCode).toContain('close[1]');
      expect(jsCode).toContain('open[1]');
    });
  });
});
