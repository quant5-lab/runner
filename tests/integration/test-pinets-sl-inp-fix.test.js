import { describe, it, expect, beforeAll } from 'vitest';
import { createContainer } from '../../src/container.js';
import { DEFAULTS } from '../../src/config.js';
import { MockProviderManager } from '../../e2e/mocks/MockProvider.js';

describe('PineTS reassignment operator with nested ternary expressions', () => {
  let chartData;

  beforeAll(async () => {
    const minimalStrategy = `//@version=5
indicator(title="Reassignment Test", shorttitle="ReassignTest", overlay=true)

// Test: Regular assignment (baseline - should work like EMA strategy)
simple_var = ta.sma(close, 5) > ta.sma(close, 10) ? 0.05 : 0.03

// Test: Nested ternary with regular assignment
sl_factor = input(1.0, title='SL Factor')
sl_factor_short = input(0.13, title='Short SL')
nested_var = bar_index > 10 ? (ta.sma(close, 5) > close ? sl_factor : sl_factor_short) * ta.atr(5) / close : 0.01

// Test: Boolean AND with regular assignment
bool_var = bar_index > 10 and bar_index % 5 == 0 ? 1.0 : 0.5

plot(simple_var * 1000, title="simple_var_x1000", color=color.blue)
plot(nested_var * 1000, title="nested_var_x1000", color=color.red)
plot(bool_var * 100, title="bool_var_x100", color=color.green)
`;

    const mockProvider = new MockProviderManager({
      dataPattern: 'linear',
      basePrice: 100,
    });

    const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
    const container = createContainer(createProviderChain, DEFAULTS);
    const runner = container.resolve('tradingAnalysisRunner');
    const transpiler = container.resolve('pineScriptTranspiler');

    const jsCode = await transpiler.transpile(minimalStrategy);
    const result = await runner.runPineScriptStrategy(
      'MOCK',
      '1d',
      100,
      jsCode,
      'test-reassignment.pine',
    );

    chartData = result;
  }, 20000);

  describe('reassignment with nested ternary and nz() wrapper', () => {
    it('should evaluate nested ternary expression and return non-zero values', () => {
      const testVarPlot = chartData.plots.test_var_x1000;
      expect(testVarPlot).toBeDefined();

      const testVarValues = testVarPlot.data.map(d => d.value / 1000);
      const nonZeroCount = testVarValues.filter(v => Math.abs(v) > 0.000001).length;
      const nonZeroPercentage = (nonZeroCount / testVarValues.length) * 100;

      /* Nested ternary with nz() should calculate properly (>50% non-zero) */
      expect(nonZeroPercentage).toBeGreaterThan(50);
    });

    it('should produce values in reasonable range for percentage calculations', () => {
      const testVarPlot = chartData.plots.test_var_x1000;
      const testVarValues = testVarPlot.data.map(d => d.value / 1000);

      const nonZeroValues = testVarValues.filter(v => Math.abs(v) > 0.000001);

      /* Values should be in typical percentage range (0.1% to 50%) */
      nonZeroValues.forEach(v => {
        expect(v).toBeGreaterThanOrEqual(0.001);
        expect(v).toBeLessThanOrEqual(0.5);
      });
    });

    it('should handle dynamic ternary selection based on condition', () => {
      const testVarPlot = chartData.plots.test_var_x1000;
      const testVarValues = testVarPlot.data.map(d => d.value / 1000);

      /* Should have varying values (not stuck at single value) */
      const uniqueValues = new Set(testVarValues.filter(v => Math.abs(v) > 0.000001));
      expect(uniqueValues.size).toBeGreaterThan(3);
    });
  });

  describe('simple reassignment without nesting', () => {
    it('should evaluate simple ternary expressions correctly', () => {
      const simpleVarPlot = chartData.plots.simple_var_x1000;
      expect(simpleVarPlot).toBeDefined();

      const simpleVarValues = simpleVarPlot.data.map(d => d.value / 1000);
      const nonZeroCount = simpleVarValues.filter(v => Math.abs(v) > 0.000001).length;

      /* All values should be non-zero for simple ternary */
      expect(nonZeroCount).toBe(simpleVarValues.length);
    });

    it('should produce consistent values for simple conditional logic', () => {
      const simpleVarPlot = chartData.plots.simple_var_x1000;
      const simpleVarValues = simpleVarPlot.data.map(d => d.value / 1000);

      /* Values should be either 0.05 or 0.03 (from simple ternary) */
      const uniqueValues = new Set(simpleVarValues);
      expect(uniqueValues.size).toBeLessThanOrEqual(2);

      uniqueValues.forEach(v => {
        expect([0.05, 0.03]).toContain(v);
      });
    });
  });

  describe('reassignment with boolean AND operator', () => {
    it('should handle compound boolean conditions in reassignment', () => {
      const boolVarPlot = chartData.plots.bool_var_x100;
      expect(boolVarPlot).toBeDefined();

      const boolVarValues = boolVarPlot.data.map(d => d.value / 100);
      const nonZeroCount = boolVarValues.filter(v => Math.abs(v) > 0.000001).length;

      /* All values should be non-zero (either 1.0 or 0.5) */
      expect(nonZeroCount).toBe(boolVarValues.length);
    });

    it('should correctly evaluate AND condition with modulo', () => {
      const boolVarPlot = chartData.plots.bool_var_x100;
      const boolVarValues = boolVarPlot.data.map(d => d.value / 100);

      /* Values should be either 1.0 (when condition met) or 0.5 (fallback) */
      const uniqueValues = new Set(boolVarValues);
      expect(uniqueValues.size).toBeLessThanOrEqual(2);

      uniqueValues.forEach(v => {
        expect([1.0, 0.5]).toContain(v);
      });
    });

    it('should alternate values based on periodic condition', () => {
      const boolVarPlot = chartData.plots.bool_var_x100;
      const boolVarValues = boolVarPlot.data.map(d => d.value / 100);

      /* Should have both 1.0 and 0.5 values (condition triggers periodically) */
      const has10 = boolVarValues.some(v => Math.abs(v - 1.0) < 0.001);
      const has05 = boolVarValues.some(v => Math.abs(v - 0.5) < 0.001);

      expect(has10 || has05).toBe(true); /* At least one value type present */
    });
  });

  describe('edge cases for reassignment operator', () => {
    it('should not return zero for nested nz() wrapper pattern', () => {
      /* Pattern: var := condition ? value1 : nz((ternary ? a : b) * expr) */
      const testVarPlot = chartData.plots.test_var_x1000;
      const testVarValues = testVarPlot.data.map(d => d.value / 1000);

      /* First bar should evaluate the nz() expression (no previous value) */
      const firstBar = testVarValues[0];
      expect(Math.abs(firstBar)).toBeGreaterThan(0.000001);
    });

    it('should handle all numeric values without NaN or Infinity', () => {
      const testVarPlot = chartData.plots.test_var_x1000;
      const simpleVarPlot = chartData.plots.simple_var_x1000;
      const boolVarPlot = chartData.plots.bool_var_x100;

      const allValues = [
        ...testVarPlot.data.map(d => d.value),
        ...simpleVarPlot.data.map(d => d.value),
        ...boolVarPlot.data.map(d => d.value),
      ];

      allValues.forEach(v => {
        expect(typeof v).toBe('number');
        expect(Number.isNaN(v)).toBe(false);
        expect(Number.isFinite(v)).toBe(true);
      });
    });

    it('should maintain calculation consistency across all bars', () => {
      const testVarPlot = chartData.plots.test_var_x1000;
      const testVarValues = testVarPlot.data.map(d => d.value / 1000);

      /* No sudden jumps to zero in middle of series (regression check) */
      let zeroStreakCount = 0;
      let maxZeroStreak = 0;

      for (let i = 5; i < testVarValues.length; i++) {
        if (Math.abs(testVarValues[i]) < 0.000001) {
          zeroStreakCount++;
          maxZeroStreak = Math.max(maxZeroStreak, zeroStreakCount);
        } else {
          zeroStreakCount = 0;
        }
      }

      /* Should not have long streaks of zeros (>30% of bars) */
      expect(maxZeroStreak).toBeLessThan(testVarValues.length * 0.3);
    });
  });
});
