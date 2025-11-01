import { createContainer } from './container.js';
import { createProviderChain, DEFAULTS } from './config.js';
import { readFile } from 'fs/promises';
import PineVersionMigrator from './pine/PineVersionMigrator.js';
import { ArgumentValidator } from './utils/argumentValidator.js';

/* Parse --settings='{"key":"value"}' from CLI arguments */
function parseSettingsArg(argv) {
  const settingsArg = argv.find((arg) => arg.startsWith('--settings='));
  if (!settingsArg) return null;

  try {
    const jsonString = settingsArg.substring('--settings='.length);
    const settings = JSON.parse(jsonString);
    if (typeof settings !== 'object' || Array.isArray(settings)) {
      throw new Error('Settings must be an object');
    }
    return settings;
  } catch (error) {
    throw new Error(`Invalid --settings format: ${error.message}`);
  }
}

async function main() {
  const startTime = performance.now();
  try {
    const { symbol, timeframe, bars } = DEFAULTS;
    
    ArgumentValidator.validateBarsArgument(process.argv[4]);
    
    const envSymbol = process.argv[2] || process.env.SYMBOL || symbol;
    const envTimeframe = process.argv[3] || process.env.TIMEFRAME || timeframe;
    const envBars = parseInt(process.argv[4]) || parseInt(process.env.BARS) || bars;
    const envStrategy = process.argv[5] || process.env.STRATEGY;
    const settings = parseSettingsArg(process.argv);

    await ArgumentValidator.validate(envSymbol, envTimeframe, envBars, envStrategy);

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

      let version = transpiler.detectVersion(pineCode);

      /* Force migration for files without @version that contain v3/v4 syntax */
      if (version === 5 && PineVersionMigrator.hasV3V4Syntax(pineCode)) {
        logger.info('v3/v4 syntax detected, applying migration');
        version = 4;
      }

      const migratedCode = PineVersionMigrator.migrate(pineCode, version);
      if (version && version < 5) {
        logger.info(`Migrated v${version} → v5`);
      }

      const transpileStartTime = performance.now();
      const jsCode = await transpiler.transpile(migratedCode);
      const transpileDuration = (performance.now() - transpileStartTime).toFixed(2);
      logger.info(`Transpilation:\ttook ${transpileDuration}ms (${jsCode.length} chars)`);

      if (settings) {
        logger.info(`Input overrides: ${JSON.stringify(settings)}`);
      }

      await runner.runPineScriptStrategy(
        envSymbol,
        envTimeframe,
        envBars,
        jsCode,
        envStrategy,
        settings,
      );

      const runDuration = (performance.now() - strategyStartTime).toFixed(2);
      logger.info(`Strategy total:\ttook ${runDuration}ms`);
    } else {
      throw new Error('No strategy file provided');
    }

    const totalDuration = (performance.now() - startTime).toFixed(2);
    logger.info(`Completed in:\ttook ${totalDuration}ms total`);

    /* Log API statistics */
    const stats = container.resolve('apiStatsCollector');
    stats.logSummary(logger);
  } catch (error) {
    const container = createContainer(createProviderChain, DEFAULTS);
    const logger = container.resolve('logger');
    logger.error('Error:', error);
    process.exit(1);
  }
}

main();
