import { createContainer } from './container.js';
import { PROVIDER_CHAIN, DEFAULTS } from './config.js';

async function main() {
  try {
    const { symbol, timeframe, bars } = DEFAULTS;
    const envSymbol = process.env.SYMBOL || symbol;
    const envTimeframe = process.env.TIMEFRAME || timeframe;
    const envBars = parseInt(process.env.BARS) || bars;

    const container = createContainer(PROVIDER_CHAIN, DEFAULTS);
    const orchestrator = container.resolve('tradingOrchestrator');

    await orchestrator.runTradingAnalysis(envSymbol, envTimeframe, envBars);
  } catch (error) {
    const container = createContainer(PROVIDER_CHAIN, DEFAULTS);
    const logger = container.resolve('logger');
    logger.error('Error:', error);
    process.exit(1);
  }
}

main();
