import { describe, it, expect, beforeEach } from 'vitest';
import { CandlestickDataSanitizer } from '../../src/classes/CandlestickDataSanitizer.js';

describe('CandlestickDataSanitizer', () => {
  let processor;

  beforeEach(() => {
    processor = new CandlestickDataSanitizer();
  });

  describe('isValidCandle()', () => {
    it('should return true for valid candle', () => {
      const candle = { open: 100, high: 105, low: 95, close: 102 };
      expect(processor.isValidCandle(candle)).toBe(true);
    });

    it('should return true for string numeric values', () => {
      const candle = { open: '100', high: '105', low: '95', close: '102' };
      expect(processor.isValidCandle(candle)).toBe(true);
    });

    it('should return false when high is not maximum', () => {
      const candle = { open: 100, high: 105, low: 95, close: 110 };
      expect(processor.isValidCandle(candle)).toBe(false);
    });

    it('should return false when low is not minimum', () => {
      const candle = { open: 100, high: 105, low: 95, close: 90 };
      expect(processor.isValidCandle(candle)).toBe(false);
    });

    it('should return false for negative values', () => {
      const candle = { open: -100, high: 105, low: 95, close: 102 };
      expect(processor.isValidCandle(candle)).toBe(false);
    });

    it('should return false for zero values', () => {
      const candle = { open: 0, high: 105, low: 95, close: 102 };
      expect(processor.isValidCandle(candle)).toBe(false);
    });

    it('should return false for NaN values', () => {
      const candle = { open: NaN, high: 105, low: 95, close: 102 };
      expect(processor.isValidCandle(candle)).toBe(false);
    });

    it('should return false for non-numeric strings', () => {
      const candle = { open: 'abc', high: 105, low: 95, close: 102 };
      expect(processor.isValidCandle(candle)).toBe(false);
    });
  });

  describe('normalizeCandle()', () => {
    it('should normalize valid candle', () => {
      const candle = {
        openTime: 1609459200000,
        open: 100,
        high: 105,
        low: 95,
        close: 102,
        volume: 5000,
      };
      const normalized = processor.normalizeCandle(candle);
      expect(normalized).toEqual({
        time: 1609459200,
        open: 100,
        high: 105,
        low: 95,
        close: 102,
        volume: 5000,
      });
    });

    it('should convert string values to numbers', () => {
      const candle = {
        openTime: 1609459200000,
        open: '100',
        high: '105',
        low: '95',
        close: '102',
        volume: '5000',
      };
      const normalized = processor.normalizeCandle(candle);
      expect(normalized.open).toBe(100);
      expect(normalized.high).toBe(105);
      expect(normalized.low).toBe(95);
      expect(normalized.close).toBe(102);
      expect(normalized.volume).toBe(5000);
    });

    it('should use default volume when missing', () => {
      const candle = {
        openTime: 1609459200000,
        open: 100,
        high: 105,
        low: 95,
        close: 102,
      };
      const normalized = processor.normalizeCandle(candle);
      expect(normalized.volume).toBe(1000);
    });

    it('should use default volume for NaN', () => {
      const candle = {
        openTime: 1609459200000,
        open: 100,
        high: 105,
        low: 95,
        close: 102,
        volume: NaN,
      };
      const normalized = processor.normalizeCandle(candle);
      expect(normalized.volume).toBe(1000);
    });

    it('should correct high to maximum of OHLC', () => {
      const candle = {
        openTime: 1609459200000,
        open: 100,
        high: 105,
        low: 95,
        close: 110,
      };
      const normalized = processor.normalizeCandle(candle);
      expect(normalized.high).toBe(110);
    });

    it('should correct low to minimum of OHLC', () => {
      const candle = {
        openTime: 1609459200000,
        open: 100,
        high: 105,
        low: 95,
        close: 90,
      };
      const normalized = processor.normalizeCandle(candle);
      expect(normalized.low).toBe(90);
    });

    it('should convert milliseconds timestamp to seconds', () => {
      const candle = {
        openTime: 1609459200123,
        open: 100,
        high: 105,
        low: 95,
        close: 102,
      };
      const normalized = processor.normalizeCandle(candle);
      expect(normalized.time).toBe(1609459200);
    });
  });

  describe('processCandlestickData()', () => {
    it('should process array of valid candles', () => {
      const rawData = [
        { openTime: 1000000, open: 100, high: 105, low: 95, close: 102, volume: 5000 },
        { openTime: 2000000, open: 102, high: 108, low: 100, close: 107, volume: 6000 },
      ];
      const processed = processor.processCandlestickData(rawData);
      expect(processed).toHaveLength(2);
      expect(processed[0].time).toBe(1000);
      expect(processed[1].time).toBe(2000);
    });

    it('should filter out invalid candles', () => {
      const rawData = [
        { openTime: 1000000, open: 100, high: 105, low: 95, close: 102 },
        { openTime: 2000000, open: 0, high: 108, low: 100, close: 107 },
        { openTime: 3000000, open: 110, high: 115, low: 108, close: 112 },
      ];
      const processed = processor.processCandlestickData(rawData);
      expect(processed).toHaveLength(2);
      expect(processed[0].open).toBe(100);
      expect(processed[1].open).toBe(110);
    });

    it('should return empty array for empty input', () => {
      const processed = processor.processCandlestickData([]);
      expect(processed).toEqual([]);
    });

    it('should return empty array for null input', () => {
      const processed = processor.processCandlestickData(null);
      expect(processed).toEqual([]);
    });

    it('should return empty array for undefined input', () => {
      const processed = processor.processCandlestickData(undefined);
      expect(processed).toEqual([]);
    });

    it('should handle single candle', () => {
      const rawData = [
        { openTime: 1000000, open: 100, high: 105, low: 95, close: 102, volume: 5000 },
      ];
      const processed = processor.processCandlestickData(rawData);
      expect(processed).toHaveLength(1);
      expect(processed[0].open).toBe(100);
    });

    it('should filter all invalid candles', () => {
      const rawData = [
        { openTime: 1000000, open: -100, high: 105, low: 95, close: 102 },
        { openTime: 2000000, open: NaN, high: 108, low: 100, close: 107 },
      ];
      const processed = processor.processCandlestickData(rawData);
      expect(processed).toEqual([]);
    });
  });
});
