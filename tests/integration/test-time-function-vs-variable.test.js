/* time() function vs time variable - Regression test for function call wrapping bug */
import { describe, it, expect, beforeAll } from 'vitest';
import { createContainer } from '../../src/container.js';
import { DEFAULTS } from '../../src/config.js';
import { MockProviderManager } from '../../e2e/mocks/MockProvider.js';

describe('time() function vs time variable', () => {
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
  it('should NOT wrap time in time() function call', async () => {
    const code = `//@version=5
indicator("Time Function", overlay=true)
x = time(timeframe.period, "0950-1645")
plot(x, "session")
`;
    
    const jsCode = await transpiler.transpile(code);
    
    // time() function call should NOT be wrapped
    expect(jsCode).toContain('time(timeframe.period');
    expect(jsCode).not.toContain('time[0](');
  });

  it('should wrap time when used as built-in variable', async () => {
    const code = `//@version=5
indicator("Time Variable", overlay=true)
t = time
plot(t, "timestamp")
`;
    
    const jsCode = await transpiler.transpile(code);
    
    // time variable should be wrapped with [0]
    expect(jsCode).toContain('time[0]');
  });

  it('should handle time() function in na() call', async () => {
    const code = `//@version=5
indicator("Session Check", overlay=true)
session_open = na(time(timeframe.period, "0950-1645")) ? false : true
plot(session_open ? 1 : 0)
`;
    
    const jsCode = await transpiler.transpile(code);
    
    // time() function should NOT be wrapped
    expect(jsCode).toContain('na(time(timeframe.period');
    expect(jsCode).not.toContain('time[0](');
  });

  it('should handle mixed time function and bar_index variable', async () => {
    const code = `//@version=5
indicator("Mixed", overlay=true)
is_session = na(time(timeframe.period, "0950-1645")) ? false : true
is_active = bar_index > 5 and is_session
plot(is_active ? 1 : 0)
`;
    
    const jsCode = await transpiler.transpile(code);
    
    // time() function NOT wrapped
    expect(jsCode).toContain('time(timeframe.period');
    expect(jsCode).not.toContain('time[0](');
    
    // bar_index variable wrapped
    expect(jsCode).toContain('bar_index[0] > 5');
  });

  it('should execute indicator with time() session filtering', async () => {
    const code = `//@version=5
indicator("Session Filter", overlay=true)
is_entry_time = na(time(timeframe.period, "0950-1345")) ? false : true
is_active = bar_index > 10 and is_entry_time
plot(is_active ? 1 : 0, "Active")
`;
    
    const result = await runner.runPineScriptStrategy('TEST', '1h', 20, code, 'test.pine');
    
    // Should execute without "time.param is not a function" error
    expect(result).toBeDefined();
    expect(result.plots).toHaveProperty('Active');
    expect(result.plots.Active.data).toHaveLength(20);
  });
});
