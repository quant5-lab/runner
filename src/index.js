import { createContainer } from './container.js';
import { createProviderChain, DEFAULTS } from './config.js';
import { readFile } from 'fs/promises';

async function main() {
  const startTime = performance.now();
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
      const strategyStartTime = performance.now();
      logger.info(`Strategy file:\t${envStrategy}`);
      const transpiler = container.resolve('pineScriptTranspiler');

      const loadStartTime = performance.now();
      const pineCode = await readFile(envStrategy, 'utf-8');
      const loadDuration = (performance.now() - loadStartTime).toFixed(2);
      logger.info(`Loading file:\ttook ${loadDuration}ms`);

      const transpileStartTime = performance.now();
      const jsCode = await transpiler.transpile(pineCode);
      const transpileDuration = (performance.now() - transpileStartTime).toFixed(2);
      logger.info(`Transpilation:\ttook ${transpileDuration}ms (${jsCode.length} chars)`);

      await runner.runPineScriptStrategy(envSymbol, envTimeframe, envBars, jsCode, envStrategy);

      const runDuration = (performance.now() - strategyStartTime).toFixed(2);
      logger.info(`Strategy total:\ttook ${runDuration}ms`);
    } else {
      await runner.runDefaultStrategy(envSymbol, envTimeframe, envBars);
    }

    const totalDuration = (performance.now() - startTime).toFixed(2);
    logger.info(`Completed in:\ttook ${totalDuration}ms total`);
  } catch (error) {
    const container = createContainer(createProviderChain, DEFAULTS);
    const logger = container.resolve('logger');
    logger.error('Error:', error);
    process.exit(1);
  }
}

main();
