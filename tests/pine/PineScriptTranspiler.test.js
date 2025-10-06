import { describe, it, expect, beforeEach, vi } from 'vitest';
import { PineScriptTranspiler, PineScriptTranspilationError } from '../../src/pine/PineScriptTranspiler.js';

describe('PineScriptTranspiler', () => {
  let transpiler;
  let mockLogger;

  beforeEach(() => {
    mockLogger = {
      info: vi.fn(),
      warn: vi.fn(),
      error: vi.fn()
    };
    transpiler = new PineScriptTranspiler(mockLogger);
  });

  describe('detectVersion()', () => {
    it('should detect version 5 from //@version=5', () => {
      const pineCode = '//@version=5\nindicator("Test")';
      const version = transpiler.detectVersion(pineCode);
      expect(version).toBe(5);
    });

    it('should detect version 4 from //@version=4', () => {
      const pineCode = '//@version=4\nstudy("Test")';
      const version = transpiler.detectVersion(pineCode);
      expect(version).toBe(4);
    });

    it('should default to version 5 when no version comment found', () => {
      const pineCode = 'indicator("Test")';
      const version = transpiler.detectVersion(pineCode);
      expect(version).toBe(5);
      expect(mockLogger.warn).toHaveBeenCalledWith(
        expect.stringContaining('No //@version comment found')
      );
    });

    it('should default to version 5 for unsupported versions', () => {
      const pineCode = '//@version=3\nindicator("Test")';
      const version = transpiler.detectVersion(pineCode);
      expect(version).toBe(5);
      expect(mockLogger.warn).toHaveBeenCalledWith(
        expect.stringContaining('Unsupported Pine Script version: 3')
      );
    });
  });

  describe('getCacheKey()', () => {
    it('should generate consistent hash for same code', () => {
      const pineCode = 'indicator("Test")\nplot(close)';
      const key1 = transpiler.getCacheKey(pineCode);
      const key2 = transpiler.getCacheKey(pineCode);
      expect(key1).toBe(key2);
      expect(key1).toMatch(/^[a-f0-9]{64}$/);
    });

    it('should generate different hashes for different code', () => {
      const code1 = 'indicator("Test1")';
      const code2 = 'indicator("Test2")';
      const key1 = transpiler.getCacheKey(code1);
      const key2 = transpiler.getCacheKey(code2);
      expect(key1).not.toBe(key2);
    });
  });

  describe('generateJavaScript()', () => {
    it('should convert ESTree AST to JavaScript', () => {
      const ast = {
        type: 'Program',
        body: [
          {
            type: 'ExpressionStatement',
            expression: {
              type: 'CallExpression',
              callee: {
                type: 'Identifier',
                name: 'indicator'
              },
              arguments: [
                {
                  type: 'Literal',
                  value: 'Test'
                }
              ]
            }
          }
        ]
      };
      const jsCode = transpiler.generateJavaScript(ast);
      expect(jsCode).toContain('indicator');
      expect(jsCode).toContain('Test');
    });

    it('should throw error for invalid AST', () => {
      const invalidAst = { invalid: 'structure' };
      expect(() => transpiler.generateJavaScript(invalidAst)).toThrow();
    });
  });
});
