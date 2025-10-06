import { createContainer } from './container.js';
import { createProviderChain, DEFAULTS } from './config.js';
import { readFile } from 'fs/promises';

async function main() {
  try {
    const { symbol, timeframe, bars } = DEFAULTS;
    const envSymbol = process.argv[2] || process.env.SYMBOL || symbol;
    const envTimeframe = process.argv[3] || process.env.TIMEFRAME || timeframe;
    const envBars = parseInt(process.argv[4]) || parseInt(process.env.BARS) || bars;
    const envStrategy = process.argv[5] || process.env.STRATEGY;

    const container = createContainer(createProviderChain, DEFAULTS);
    const logger = container.resolve('logger');
    const runner = container.resolve('tradingAnalysisRunner');

    if (envStrategy) {
      logger.info(`🌲 Pine Script strategy file: ${envStrategy}`);
      const transpiler = container.resolve('pineScriptTranspiler');

      const pineCode = await readFile(envStrategy, 'utf-8');
      logger.info('📖 Pine Script code loaded, transpiling...');

      const jsCode = await transpiler.transpile(pineCode);
      logger.info('✅ Transpilation complete, generated JavaScript');
      logger.info(`📝 Transpiled code length: ${jsCode.length} characters`);

      await runner.run(envSymbol, envTimeframe, envBars, jsCode);
    } else {
      await runner.run(envSymbol, envTimeframe, envBars);
    }
  } catch (error) {
    const container = createContainer(createProviderChain, DEFAULTS);
    const logger = container.resolve('logger');
    logger.error('Error:', error);
    process.exit(1);
  }
}

main();
