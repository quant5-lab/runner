import { TimeframeError } from '../errors/TimeframeError.js';

/* Shared constants: Supported timeframes in unified app format (DRY principle) */
/* Unified format: D (daily), W (weekly), M (monthly), xh (hourly), xm (minute) */
export const SUPPORTED_TIMEFRAMES = {
  MOEX: ['1m', '10m', '1h', 'D', 'W', 'M'],
  BINANCE: ['1m', '3m', '5m', '15m', '30m', '1h', '2h', '4h', '6h', '8h', '12h', 'D', '3d', 'W', 'M'],
  YAHOO: ['1m', '2m', '5m', '15m', '30m', '1h', '90m', 'D', 'W', 'M'],
};

/**
 * Utility to parse timeframe strings into standardized formats
 */
export class TimeframeParser {
  /**
   * Parse timeframe string (e.g., "15m", "1h", "D") into minutes
   * Supports unified format (D, W, M) and legacy formats (1d, 1w, 1M, 1wk, 1mo)
   * @param {string|number} timeframe - Input timeframe
   * @returns {number} - Timeframe in minutes
   */
  static parseToMinutes(timeframe) {
    if (typeof timeframe === 'number') {
      return timeframe;
    }

    const str = String(timeframe);

    /* Normalize legacy formats to unified format for backward compatibility */
    const normalized = str
      .replace(/^1d$/i, 'D')      // 1d → D (daily)
      .replace(/^1wk$/i, 'W')     // 1wk → W (weekly, Yahoo legacy)
      .replace(/^1w$/i, 'W')      // 1w → W (weekly, provider legacy)
      .replace(/^1mo$/i, 'M');    // 1mo → M (monthly, Yahoo legacy)
      // Note: 1M stays as-is, handled in next section

    // Handle unified letter formats - D, W, M don't support digit prefixes
    if (normalized === 'D') return 1440; // Daily = 1440 minutes
    if (normalized === 'W') return 10080; // Weekly = 7 * 1440 minutes
    if (normalized === 'M' || normalized === '1M') return 43200; // Monthly = 30 * 1440 minutes

    // Parse number + unit format (e.g., "15m", "1h")
    const match = normalized.match(/^(\d+)([mh])$/);
    if (!match) {
      return 1440; // Default to daily if can't parse
    }

    const [, value, unit] = match;
    const num = parseInt(value, 10);

    switch (unit) {
      case 'm': return num; // minutes
      case 'h': return num * 60; // hours to minutes
      default: return 1440; // Default to daily
    }
  }

  /**
   * Generic conversion helper - DRY pattern
   * @param {string|number} timeframe - Input timeframe
   * @param {Object} mapping - Minutes to provider format mapping
   * @param {string} providerName - Provider name for error messages
   * @param {Array} supportedTimeframes - Supported timeframes list
   * @returns {string} - Converted timeframe
   * @private
   */
  static _convertTimeframe(timeframe, mapping, providerName, supportedTimeframes) {
    const minutes = this.parseToMinutes(timeframe);
    const converted = mapping[minutes];
    if (converted === undefined) {
      throw new TimeframeError(timeframe, providerName, supportedTimeframes);
    }
    return converted;
  }

  /**
   * Convert timeframe to MOEX interval format
   * @param {string|number} timeframe - Input timeframe
   * @returns {string} - MOEX interval
   */
  static toMoexInterval(timeframe) {
    /* MOEX specific mapping - based on actual API testing */
    const mapping = {
      1: '1', 10: '10', 60: '60',
      1440: '24', 10080: '7', 43200: '31',
    };
    return this._convertTimeframe(timeframe, mapping, 'MOEX', SUPPORTED_TIMEFRAMES.MOEX);
  }

  /**
   * Convert timeframe to Yahoo Finance interval format
   * @param {string|number} timeframe - Input timeframe
   * @returns {string} - Yahoo Finance interval
   */
  static toYahooInterval(timeframe) {
    /* Yahoo Finance specific mapping */
    const mapping = {
      1: '1m', 2: '2m', 5: '5m', 15: '15m', 30: '30m',
      60: '1h', 90: '90m',
      1440: '1d', 10080: '1wk', 43200: '1mo',
    };
    return this._convertTimeframe(timeframe, mapping, 'Yahoo Finance', SUPPORTED_TIMEFRAMES.YAHOO);
  }

  /**
   * Convert timeframe to Binance format (numeric strings and letters)
   * @param {string|number} timeframe - Input timeframe
   * @returns {string} - Binance timeframe format
   */
  static toBinanceTimeframe(timeframe) {
    /* Binance expects numeric strings for most timeframes */
    const mapping = {
      1: '1', 3: '3', 5: '5', 15: '15', 30: '30',
      60: '60', 120: '120', 240: '240', 360: '360', 480: '480', 720: '720',
      1440: 'D', 10080: 'W', 43200: 'M',
    };
    return this._convertTimeframe(timeframe, mapping, 'Binance', SUPPORTED_TIMEFRAMES.BINANCE);
  }
}
