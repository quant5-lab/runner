import { describe, it, expect, vi, beforeEach } from 'vitest';
import { Logger } from '../../src/classes/Logger.js';

describe('Logger', () => {
  let logger;
  let consoleLogSpy;
  let consoleErrorSpy;

  beforeEach(() => {
    logger = new Logger();
    consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  describe('log()', () => {
    it('should call console.log with message', () => {
      logger.log('Test message');
      expect(consoleLogSpy).toHaveBeenCalledWith('Test message');
      expect(consoleLogSpy).toHaveBeenCalledTimes(1);
    });

    it('should handle empty string', () => {
      logger.log('');
      expect(consoleLogSpy).toHaveBeenCalledWith('');
    });

    it('should handle objects', () => {
      const obj = { key: 'value' };
      logger.log(obj);
      expect(consoleLogSpy).toHaveBeenCalledWith(obj);
    });
  });

  describe('error()', () => {
    it('should call console.error with single argument', () => {
      logger.error('Error message');
      expect(consoleErrorSpy).toHaveBeenCalledWith('Error message');
      expect(consoleErrorSpy).toHaveBeenCalledTimes(1);
    });

    it('should handle multiple arguments', () => {
      logger.error('Error:', 'message', 123);
      expect(consoleErrorSpy).toHaveBeenCalledWith('Error:', 'message', 123);
    });

    it('should handle Error objects', () => {
      const error = new Error('Test error');
      logger.error(error);
      expect(consoleErrorSpy).toHaveBeenCalledWith(error);
    });

    it('should handle no arguments', () => {
      logger.error();
      expect(consoleErrorSpy).toHaveBeenCalledWith();
    });
  });
});
