import { describe, it, expect } from 'vitest';
import { createContainer } from '../../src/container.js';
import { readFile } from 'fs/promises';
import { DEFAULTS } from '../../src/config.js';
import { MockProviderManager } from '../../e2e/mocks/MockProvider.js';

/* Integration test: ema-strategy.pine produces valid plots with correct EMA calculations */
describe('EMA Strategy Integration', () => {
  it('should produce EMA 1, EMA 2, and Bull Signal plots', async () => {
    const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
    const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
    const container = createContainer(createProviderChain, DEFAULTS);
    const runner = container.resolve('tradingAnalysisRunner');
    const transpiler = container.resolve('pineScriptTranspiler');

    const pineCode = await readFile('strategies/ema-strategy.pine', 'utf-8');
    const jsCode = await transpiler.transpile(pineCode);

    const result = await runner.runPineScriptStrategy(
      'BTCUSDT',
      '1h',
      100,
      jsCode,
      'strategies/ema-strategy.pine',
    );

    expect(result.plots).toBeDefined();
    expect(result.plots['EMA 1']).toBeDefined();
    expect(result.plots['EMA 2']).toBeDefined();
    expect(result.plots['Bull Signal']).toBeDefined();

    const ema1Data = result.plots['EMA 1'].data;
    const ema2Data = result.plots['EMA 2'].data;
    const bullSignalData = result.plots['Bull Signal'].data;

    expect(ema1Data.length).toBeGreaterThan(0);
    expect(ema2Data.length).toBeGreaterThan(0);
    expect(bullSignalData.length).toBeGreaterThan(0);

    /* Verify EMA 1 values (20-period needs 19 warmup bars) */
    const validEma1 = ema1Data.filter((d) => typeof d.value === 'number' && !isNaN(d.value));
    expect(validEma1.length).toBeGreaterThanOrEqual(ema1Data.length - 20);

    /* Verify EMA 2 values (50-period needs 49 warmup bars) */
    const validEma2 = ema2Data.filter((d) => typeof d.value === 'number' && !isNaN(d.value));
    expect(validEma2.length).toBeGreaterThanOrEqual(ema2Data.length - 50);

    /* Verify Bull Signal is 0 or 1 */
    bullSignalData.forEach((d, i) => {
      expect([0, 1]).toContain(d.value);
    });
  });

  it('should calculate Bull Signal correctly (1 when EMA1 > EMA2, 0 otherwise)', async () => {
    const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
    const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
    const container = createContainer(createProviderChain, DEFAULTS);
    const runner = container.resolve('tradingAnalysisRunner');
    const transpiler = container.resolve('pineScriptTranspiler');

    const pineCode = await readFile('strategies/ema-strategy.pine', 'utf-8');
    const jsCode = await transpiler.transpile(pineCode);

    const result = await runner.runPineScriptStrategy(
      'BTCUSDT',
      '1h',
      100,
      jsCode,
      'strategies/ema-strategy.pine',
    );

    const ema1Data = result.plots['EMA 1'].data;
    const ema2Data = result.plots['EMA 2'].data;
    const bullSignalData = result.plots['Bull Signal'].data;

    /* Compare Bull Signal logic */
    for (let i = 0; i < bullSignalData.length; i++) {
      const ema1 = ema1Data[i]?.value;
      const ema2 = ema2Data[i]?.value;
      const bullSignal = bullSignalData[i]?.value;

      if (typeof ema1 === 'number' && typeof ema2 === 'number' && !isNaN(ema1) && !isNaN(ema2)) {
        const expectedSignal = ema1 > ema2 ? 1 : 0;
        expect(bullSignal).toBe(expectedSignal);
      }
    }
  });

  it('should calculate EMA 1 (20-period) correctly', async () => {
    const mockProvider = new MockProviderManager({ dataPattern: 'linear', basePrice: 100 });
    const createProviderChain = () => [{ name: 'MockProvider', instance: mockProvider }];
    const container = createContainer(createProviderChain, DEFAULTS);
    const runner = container.resolve('tradingAnalysisRunner');
    const transpiler = container.resolve('pineScriptTranspiler');

    const pineCode = await readFile('strategies/ema-strategy.pine', 'utf-8');
    const jsCode = await transpiler.transpile(pineCode);

    const result = await runner.runPineScriptStrategy(
      'BTCUSDT',
      '1h',
      100,
      jsCode,
      'strategies/ema-strategy.pine',
    );

    const ema1Data = result.plots['EMA 1'].data;

    /* Manual EMA calculation verification for first few values */
    const providerManager = container.resolve('providerManager');
    const { data: marketData } = await providerManager.fetchMarketData('BTCUSDT', '1h', 100);

    const closes = marketData.map((candle) => candle.close);
    const period = 20;
    const multiplier = 2 / (period + 1);

    /* Calculate EMA manually */
    let ema = closes[0]; // Start with first close as initial EMA
    const manualEma = [ema];

    for (let i = 1; i < closes.length; i++) {
      ema = closes[i] * multiplier + ema * (1 - multiplier);
      manualEma.push(ema);
    }

    /* Compare last 10 values (most stable) */
    const lastN = 10;
    for (let i = ema1Data.length - lastN; i < ema1Data.length; i++) {
      const plotValue = ema1Data[i].value;
      const expectedValue = manualEma[i];
      const tolerance = expectedValue * 0.01; // 1% tolerance
      expect(Math.abs(plotValue - expectedValue)).toBeLessThan(tolerance);
    }
  });
});
