import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { writeFile, unlink, mkdir } from 'fs/promises';
import { ArgumentValidator } from '../../src/utils/argumentValidator.js';

describe('ArgumentValidator', () => {
  describe('validateSymbol', () => {
    it('should accept valid symbol', () => {
      expect(() => ArgumentValidator.validateSymbol('BTCUSDT')).not.toThrow();
      expect(() => ArgumentValidator.validateSymbol('AAPL')).not.toThrow();
    });

    it('should reject empty string', () => {
      expect(() => ArgumentValidator.validateSymbol('')).toThrow('Symbol must be a non-empty string');
    });

    it('should reject whitespace-only string', () => {
      expect(() => ArgumentValidator.validateSymbol('   ')).toThrow('Symbol must be a non-empty string');
    });

    it('should reject non-string', () => {
      expect(() => ArgumentValidator.validateSymbol(null)).toThrow('Symbol must be a non-empty string');
      expect(() => ArgumentValidator.validateSymbol(undefined)).toThrow('Symbol must be a non-empty string');
      expect(() => ArgumentValidator.validateSymbol(123)).toThrow('Symbol must be a non-empty string');
    });
  });

  describe('validateTimeframe', () => {
    it('should accept valid timeframes', () => {
      expect(() => ArgumentValidator.validateTimeframe('1h')).not.toThrow();
      expect(() => ArgumentValidator.validateTimeframe('D')).not.toThrow();
      expect(() => ArgumentValidator.validateTimeframe('1d')).not.toThrow();
      expect(() => ArgumentValidator.validateTimeframe('M')).not.toThrow();
    });

    it('should reject invalid timeframe', () => {
      expect(() => ArgumentValidator.validateTimeframe('INVALID')).toThrow('Timeframe must be one of:');
      expect(() => ArgumentValidator.validateTimeframe('2D')).toThrow('Timeframe must be one of:');
    });

    it('should reject empty/null timeframe', () => {
      expect(() => ArgumentValidator.validateTimeframe('')).toThrow('Timeframe must be one of:');
      expect(() => ArgumentValidator.validateTimeframe(null)).toThrow('Timeframe must be one of:');
    });
  });

  describe('validateBars', () => {
    it('should accept valid bars count', () => {
      expect(() => ArgumentValidator.validateBars(1)).not.toThrow();
      expect(() => ArgumentValidator.validateBars(100)).not.toThrow();
      expect(() => ArgumentValidator.validateBars(5000)).not.toThrow();
    });

    it('should reject bars below minimum', () => {
      expect(() => ArgumentValidator.validateBars(0)).toThrow('Bars must be a number between 1 and 5000');
      expect(() => ArgumentValidator.validateBars(-10)).toThrow('Bars must be a number between 1 and 5000');
    });

    it('should reject bars above maximum', () => {
      expect(() => ArgumentValidator.validateBars(5001)).toThrow('Bars must be a number between 1 and 5000');
      expect(() => ArgumentValidator.validateBars(10000)).toThrow('Bars must be a number between 1 and 5000');
    });

    it('should reject NaN', () => {
      expect(() => ArgumentValidator.validateBars(NaN)).toThrow('Bars must be a number between 1 and 5000');
    });
  });

  describe('validateBarsArgument', () => {
    it('should accept numeric string', () => {
      expect(() => ArgumentValidator.validateBarsArgument('100')).not.toThrow();
      expect(() => ArgumentValidator.validateBarsArgument('1')).not.toThrow();
      expect(() => ArgumentValidator.validateBarsArgument('5000')).not.toThrow();
    });

    it('should accept undefined', () => {
      expect(() => ArgumentValidator.validateBarsArgument(undefined)).not.toThrow();
    });

    it('should reject non-numeric string', () => {
      expect(() => ArgumentValidator.validateBarsArgument('strategies/test.pine')).toThrow('Bars must be a number');
      expect(() => ArgumentValidator.validateBarsArgument('abc')).toThrow('Bars must be a number');
      expect(() => ArgumentValidator.validateBarsArgument('100.5')).toThrow('Bars must be a number');
    });
  });

  describe('validateStrategyFile', () => {
    const testDir = '/tmp/test-strategies';
    const testFile = `${testDir}/test.pine`;

    beforeEach(async () => {
      await mkdir(testDir, { recursive: true });
      await writeFile(testFile, 'strategy.entry("test", strategy.long)');
    });

    afterEach(async () => {
      try {
        await unlink(testFile);
      } catch {
        /* Ignore cleanup errors */
      }
    });

    it('should accept undefined strategy', async () => {
      await expect(ArgumentValidator.validateStrategyFile(undefined)).resolves.not.toThrow();
    });

    it('should accept valid .pine file', async () => {
      await expect(ArgumentValidator.validateStrategyFile(testFile)).resolves.not.toThrow();
    });

    it('should reject non-.pine extension', async () => {
      await expect(ArgumentValidator.validateStrategyFile('test.js')).rejects.toThrow('Strategy file must have .pine extension');
    });

    it('should reject non-existent file', async () => {
      await expect(ArgumentValidator.validateStrategyFile('/nonexistent/test.pine')).rejects.toThrow('Strategy file not found or not readable');
    });
  });

  describe('validate', () => {
    const testDir = '/tmp/test-strategies';
    const testFile = `${testDir}/test.pine`;

    beforeEach(async () => {
      await mkdir(testDir, { recursive: true });
      await writeFile(testFile, 'strategy.entry("test", strategy.long)');
    });

    afterEach(async () => {
      try {
        await unlink(testFile);
      } catch {
        /* Ignore cleanup errors */
      }
    });

    it('should accept valid arguments', async () => {
      await expect(ArgumentValidator.validate('BTCUSDT', '1h', 100, testFile)).resolves.not.toThrow();
    });

    it('should accept valid arguments without strategy', async () => {
      await expect(ArgumentValidator.validate('BTCUSDT', '1h', 100, undefined)).resolves.not.toThrow();
    });

    it('should reject multiple invalid arguments', async () => {
      try {
        await ArgumentValidator.validate('', 'INVALID', 0, 'test.js');
        expect.fail('Should have thrown error');
      } catch (error) {
        expect(error.message).toContain('Invalid arguments:');
        expect(error.message).toContain('Symbol must be a non-empty string');
        expect(error.message).toContain('Timeframe must be one of:');
        expect(error.message).toContain('Bars must be a number between 1 and 5000');
        expect(error.message).toContain('Strategy file must have .pine extension');
      }
    });
  });
});
