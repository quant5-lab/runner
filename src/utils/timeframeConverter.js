import { TimeframeError } from '../errors/TimeframeError.js';
import { SUPPORTED_TIMEFRAMES } from './timeframeParser.js';

/* Unified timeframe format converter - DRY/SRP compliant */
class TimeframeConverter {
  /**
   * Convert minutes to PineTS format
   * @param {number} minutes - Timeframe in minutes
   * @returns {string} - PineTS format
   */
  static toPineTS(minutes) {
    const mapping = {
      1: '1',
      3: '3',
      5: '5',
      15: '15',
      30: '30',
      45: '45',
      60: '60',
      120: '120',
      180: '180',
      240: '240',
      1440: 'D',
      10080: 'W',
      43200: 'M',
    };
    return mapping[minutes] || String(minutes);
  }

  /**
   * Convert PineTS format to app timeframe string
   * @param {string} pineTF - PineTS format
   * @returns {string} - App timeframe format (e.g., '10m', '1h', 'D')
   */
  static fromPineTS(pineTF) {
    const mapping = {
      1: '1m',
      3: '3m',
      5: '5m',
      15: '15m',
      30: '30m',
      45: '45m',
      60: '1h',
      120: '2h',
      180: '3h',
      240: '4h',
      D: 'D',
      W: 'W',
      M: 'M',
    };
    /* Fallback: assume numeric string is minutes */
    return mapping[pineTF] || `${pineTF}m`;
  }

  /**
   * Convert minutes to MOEX interval format
   * @param {number} minutes - Timeframe in minutes
   * @param {string|number} originalTimeframe - Original input for error messages
   * @returns {string} - MOEX interval
   */
  static toMoex(minutes, originalTimeframe = minutes) {
    const mapping = {
      1: '1',
      10: '10',
      60: '60',
      1440: '24',
      10080: '7',
      43200: '31',
    };
    if (mapping[minutes] === undefined) {
      throw new TimeframeError(originalTimeframe, 'MOEX', SUPPORTED_TIMEFRAMES.MOEX);
    }
    return mapping[minutes];
  }

  /**
   * Convert minutes to Yahoo Finance interval format
   * @param {number} minutes - Timeframe in minutes
   * @param {string|number} originalTimeframe - Original input for error messages
   * @returns {string} - Yahoo Finance interval
   */
  static toYahoo(minutes, originalTimeframe = minutes) {
    const mapping = {
      1: '1m',
      2: '2m',
      5: '5m',
      15: '15m',
      30: '30m',
      60: '1h',
      90: '90m',
      1440: '1d',
      10080: '1wk',
      43200: '1mo',
    };
    if (mapping[minutes] === undefined) {
      throw new TimeframeError(originalTimeframe, 'Yahoo Finance', SUPPORTED_TIMEFRAMES.YAHOO);
    }
    return mapping[minutes];
  }

  /**
   * Convert minutes to Binance timeframe format
   * @param {number} minutes - Timeframe in minutes
   * @param {string|number} originalTimeframe - Original input for error messages
   * @returns {string} - Binance timeframe format
   */
  static toBinance(minutes, originalTimeframe = minutes) {
    const mapping = {
      1: '1',
      3: '3',
      5: '5',
      15: '15',
      30: '30',
      60: '60',
      120: '120',
      240: '240',
      360: '360',
      480: '480',
      720: '720',
      1440: 'D',
      10080: 'W',
      43200: 'M',
    };
    if (mapping[minutes] === undefined) {
      throw new TimeframeError(originalTimeframe, 'Binance', SUPPORTED_TIMEFRAMES.BINANCE);
    }
    return mapping[minutes];
  }
}

export default TimeframeConverter;
