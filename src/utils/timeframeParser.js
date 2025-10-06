import { TimeframeError } from '../errors/TimeframeError.js';

/**
 * Utility to parse timeframe strings into standardized formats
 */
export class TimeframeParser {
  /**
   * Parse timeframe string (e.g., "15m", "1h", "1d") into minutes
   * @param {string|number} timeframe - Input timeframe
   * @returns {number} - Timeframe in minutes
   */
  static parseToMinutes(timeframe) {
    if (typeof timeframe === 'number') {
      return timeframe;
    }

    const str = String(timeframe);

    // Handle simple letter formats - D, W, M don't support digit prefixes
    if (str === 'D') return 1440; // Daily = 1440 minutes
    if (str === 'W') return 10080; // Weekly = 7 * 1440 minutes
    if (str === 'M') return 43200; // Monthly = 30 * 1440 minutes

    // Parse number + unit format (e.g., "15m", "1h", "1d")
    // Note: "1w", "1M", "1W" etc. are INVALID - large timeframes don't support digits
    const match = str.match(/^(\d+)([mhd])$/);
    if (!match) {
      return 1440; // Default to daily if can't parse
    }

    const [, value, unit] = match;
    const num = parseInt(value, 10);

    switch (unit) {
      case 'm': return num; // minutes
      case 'h': return num * 60; // hours to minutes
      case 'd': return num * 1440; // days to minutes
      default: return 1440; // Default to daily
    }
  }

  /**
   * Convert timeframe to MOEX interval format
   * @param {string|number} timeframe - Input timeframe
   * @returns {string} - MOEX interval
   */
  static toMoexInterval(timeframe) {
    const minutes = this.parseToMinutes(timeframe);

    // MOEX specific mapping - based on actual API testing
    const mapping = {
      1: '1', // 1 minute
      10: '10', // 10 minutes
      60: '60', // 1 hour
      1440: '24', // Daily
      10080: '7', // Weekly (not exact, but closest)
      43200: '31', // Monthly (approximate)
    };

    const moexInterval = mapping[minutes];
    if (moexInterval === undefined) {
      throw new TimeframeError(timeframe, 'MOEX', '1m,10m,1h,1d,1w,1M');
    }

    return moexInterval;
  }

  /**
   * Convert timeframe to Yahoo Finance interval format
   * @param {string|number} timeframe - Input timeframe
   * @returns {string} - Yahoo Finance interval
   */
  static toYahooInterval(timeframe) {
    const minutes = this.parseToMinutes(timeframe);

    // Yahoo Finance specific mapping
    const mapping = {
      1: '1m', // 1 minute
      2: '2m', // 2 minutes
      5: '5m', // 5 minutes
      15: '15m', // 15 minutes
      30: '30m', // 30 minutes
      60: '1h', // 1 hour
      90: '90m', // 90 minutes
      1440: '1d', // Daily
      10080: '1wk', // Weekly
      43200: '1mo', // Monthly
    };

    const yahooInterval = mapping[minutes];
    if (yahooInterval === undefined) {
      throw new TimeframeError(timeframe, 'Yahoo Finance', '1m,2m,5m,15m,30m,1h,90m,1d,1wk,1mo');
    }

    return yahooInterval;
  }

  /**
   * Convert timeframe to Binance format (numeric strings and letters)
   * @param {string|number} timeframe - Input timeframe
   * @returns {string} - Binance timeframe format
   */
  static toBinanceTimeframe(timeframe) {
    const minutes = this.parseToMinutes(timeframe);

    // Binance expects numeric strings for most timeframes
    // Based on timeframe_to_binance mapping in PineTS
    const mapping = {
      1: '1', // 1 minute -> "1"
      3: '3', // 3 minutes -> "3"
      5: '5', // 5 minutes -> "5"
      15: '15', // 15 minutes -> "15"
      30: '30', // 30 minutes -> "30"
      60: '60', // 1 hour -> "60"
      120: '120', // 2 hours -> "120"
      240: '240', // 4 hours -> "240"
      360: '360', // 6 hours -> "360"
      480: '480', // 8 hours -> "480"
      720: '720', // 12 hours -> "720"
      1440: 'D', // Daily -> "D"
      10080: 'W', // Weekly -> "W"
      43200: 'M', // Monthly -> "M"
    };

    const binanceTimeframe = mapping[minutes];
    if (binanceTimeframe === undefined) {
      throw new TimeframeError(timeframe, 'Binance', '1m,3m,5m,15m,30m,1h,2h,4h,6h,8h,12h,1d,1w,1M');
    }

    return binanceTimeframe;
  }
}
