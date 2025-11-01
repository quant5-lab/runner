import { access, constants } from 'fs/promises';
import { VALID_INPUT_TIMEFRAMES } from './timeframeParser.js';

const MIN_BARS = 1;
const MAX_BARS = 5000;

export class ArgumentValidator {
  static validateSymbol(symbol) {
    if (!symbol || typeof symbol !== 'string' || symbol.trim().length === 0) {
      throw new Error('Symbol must be a non-empty string');
    }
  }

  static validateTimeframe(timeframe) {
    if (!timeframe || !VALID_INPUT_TIMEFRAMES.includes(timeframe)) {
      throw new Error(`Timeframe must be one of: ${VALID_INPUT_TIMEFRAMES.join(', ')}`);
    }
  }

  static validateBars(bars) {
    if (isNaN(bars) || bars < MIN_BARS || bars > MAX_BARS) {
      throw new Error(`Bars must be a number between ${MIN_BARS} and ${MAX_BARS}`);
    }
  }

  static validateBarsArgument(barsArg) {
    if (barsArg && !/^\d+$/.test(barsArg)) {
      throw new Error(`Argument 4 (bars) must be a number, got: "${barsArg}". Usage: node src/index.js SYMBOL TIMEFRAME BARS STRATEGY`);
    }
  }

  static async validateStrategyFile(strategyPath) {
    if (!strategyPath) return;
    
    if (!strategyPath.endsWith('.pine')) {
      throw new Error('Strategy file must have .pine extension');
    }
    
    try {
      await access(strategyPath, constants.R_OK);
    } catch {
      throw new Error(`Strategy file not found or not readable: ${strategyPath}`);
    }
  }

  static async validate(symbol, timeframe, bars, strategyPath) {
    const errors = [];

    try { this.validateSymbol(symbol); } catch (e) { errors.push(e.message); }
    try { this.validateTimeframe(timeframe); } catch (e) { errors.push(e.message); }
    try { this.validateBars(bars); } catch (e) { errors.push(e.message); }
    try { await this.validateStrategyFile(strategyPath); } catch (e) { errors.push(e.message); }

    if (errors.length > 0) {
      throw new Error(`Invalid arguments:\n  - ${errors.join('\n  - ')}`);
    }
  }
}
