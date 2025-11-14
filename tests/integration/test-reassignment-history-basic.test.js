import { describe, it, expect, beforeAll } from 'vitest';
import { createContainer } from '../../src/container.js';
import { DEFAULTS } from '../../src/config.js';
import { MockProviderManager } from '../../e2e/mocks/MockProvider.js';

/* Test parser fix: reassignment with history reference initialization */
describe('Reassignment Operator: History Reference Initialization', () => {
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

  it('should initialize variable before history access in simple case', async () => {
    const code = `//@version=5
indicator("Simple History Test", overlay=true)

sl := close > open ? sl[1] : close * 0.05
plot(sl * 1000, "sl_x1000")
`;

    const jsCode = await transpiler.transpile(code);

    /* Verify transpiler generates initialization */
    expect(jsCode).toMatch(/let sl = 0/);
    expect(jsCode).toContain('sl =');

    const result = await runner.runPineScriptStrategy('TEST', '1h', 20, jsCode, 'simple-history.pine');

    expect(result.plots.sl_x1000).toBeDefined();
    const slValues = result.plots.sl_x1000.data.map(d => d.value / 1000);

    /* All values should be calculated correctly */
    for (let i = 0; i < slValues.length; i++) {
      expect(slValues[i]).toBeGreaterThan(0);
      expect(slValues[i]).not.toBeNaN();
    }
  }, 15000);

  it('should handle conditional history reference pattern', async () => {
    const code = `//@version=5
indicator("Conditional History Test", overlay=true)

active = bar_index > 5 and bar_index < 15
value := active ? value[1] : close * 0.02
plot(value * 1000, "value_x1000")
plot(active ? 1 : 0, "active")
`;

    const jsCode = await transpiler.transpile(code);

    /* Add bar_index plot for debugging */
    const debugCode = jsCode.replace(/plot\(active \? 1 : 0, "active"\);?/, 'plot(active ? 1 : 0, "active");\nplot(bar_index, "bar_idx");');

    const result = await runner.runPineScriptStrategy('TEST', '1h', 20, debugCode, 'conditional-history.pine');

    const values = result.plots.value_x1000.data.map(d => d.value / 1000);
    const active = result.plots.active.data.map(d => d.value);

    const debug = {
      bar5: { active: active[5], value: values[5] },
      bar6: { active: active[6], value: values[6] },
      bar7: { active: active[7], value: values[7] },
      bar14: { active: active[14], value: values[14] },
      bar15: { active: active[15], value: values[15] },
    };
    require('fs').writeFileSync('/tmp/test2-debug.json', JSON.stringify(debug, null, 2));

    /* Verify preservation when active */
    expect(active[6]).toBe(1);
    const preserved = values[6];
    expect(preserved).toBeGreaterThan(0);

    for (let i = 7; i < 15; i++) {
      expect(Math.abs(values[i] - preserved)).toBeLessThan(0.0001);
    }

    /* Verify recalculation when inactive */
    expect(active[15]).toBe(0);
    expect(Math.abs(values[15] - preserved)).toBeGreaterThan(0.001);
  }, 15000);

  /* PineTS 6b82b42: No ParamMarker crash but values are 0 (initialization or evaluation broken) */
  it('should handle multiple variables with history references', async () => {
    const code = `//@version=5
indicator("Multiple History Test", overlay=true)

condition = bar_index > 5
var1 := condition ? var1[1] : close * 0.01
var2 := condition ? var2[1] : close * 0.02
var3 := condition ? var3[1] : close * 0.03

plot(var1 * 1000, "var1")
plot(var2 * 1000, "var2")
plot(var3 * 1000, "var3")
`;

    const jsCode = await transpiler.transpile(code);

    /* Verify all three variables initialized */
    expect(jsCode).toContain('let var1 = 0');
    expect(jsCode).toContain('let var2 = 0');
    expect(jsCode).toContain('let var3 = 0');

    const result = await runner.runPineScriptStrategy('TEST', '1h', 15, jsCode, 'multiple-history.pine');

    const var1 = result.plots.var1.data.map(d => d.value / 1000);
    const var2 = result.plots.var2.data.map(d => d.value / 1000);
    const var3 = result.plots.var3.data.map(d => d.value / 1000);

    /* All variables should have valid values */
    for (let i = 0; i < 15; i++) {
      expect(var1[i]).toBeGreaterThan(0);
      expect(var2[i]).toBeGreaterThan(0);
      expect(var3[i]).toBeGreaterThan(0);
    }

    /* Variables should maintain different proportions */
    expect(var2[10]).toBeCloseTo(var1[10] * 2, 1);
    expect(var3[10]).toBeCloseTo(var1[10] * 3, 1);
  }, 15000);

  /* PineTS 6b82b42: No ParamMarker crash but values are 0 (initialization or evaluation broken) */
  it('should handle nested ternary with history reference', async () => {
    const code = `//@version=5
indicator("Nested Ternary Test", overlay=true)

trigger1 = bar_index > 5
trigger2 = bar_index > 10
value := trigger1 ? (trigger2 ? value[1] : close * 0.02) : close * 0.01
plot(value * 1000, "value")
`;

    const jsCode = await transpiler.transpile(code);

    /* Verify initialization */
    expect(jsCode).toContain('let value = 0');

    const result = await runner.runPineScriptStrategy('TEST', '1h', 20, jsCode, 'nested-ternary-history.pine');

    const values = result.plots.value.data.map(d => d.value / 1000);

    /* All values should be valid */
    for (let i = 0; i < 20; i++) {
      expect(values[i]).toBeGreaterThan(0);
      expect(values[i]).not.toBeNaN();
    }
  }, 15000);
});
