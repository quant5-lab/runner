import { describe, it, expect, beforeEach, vi } from 'vitest';
import { PineScriptTranspiler } from '../../src/pine/PineScriptTranspiler.js';
import PineVersionMigrator from '../../src/pine/PineVersionMigrator.js';

/* Tests in "Full Migration + Transpilation Sequence" call real Python parser subprocess */
const TRANSPILER_TIMEOUT = 10000;

describe('PineScriptTranspiler', () => {
  let transpiler;
  let mockLogger;

  beforeEach(() => {
    mockLogger = {
      info: vi.fn(),
      warn: vi.fn(),
      error: vi.fn(),
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
    });

    it('should return actual version for all versions', () => {
      const pineCode = '//@version=3\nindicator("Test")';
      const version = transpiler.detectVersion(pineCode);
      expect(version).toBe(3);
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
                name: 'indicator',
              },
              arguments: [
                {
                  type: 'Literal',
                  value: 'Test',
                },
              ],
            },
          },
        ],
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

  describe('Full Migration + Transpilation Sequence', () => {
    it('should transform v4 input.integer through full pipeline to input.int()', async () => {
      // Stage 1: v4 source code with input.integer
      const v4Code = `//@version=4
indicator("Test")
max_trades = input(1, title='Max Trades', type=input.integer)
sl_factor = input(1.5, title='SL Factor', type=input.float)
show_trades = input(true, title='Show', type=input.bool)`;

      // Stage 2: Migrate v4 → v5
      const migratedCode = PineVersionMigrator.migrate(v4Code, 4);
      expect(migratedCode).toContain('type=input.int)');
      expect(migratedCode).not.toContain('type=input.integer)');

      // Stage 3: Transpile to JavaScript
      const jsCode = await transpiler.transpile(migratedCode);
      expect(jsCode).toBeDefined();
      expect(typeof jsCode).toBe('string');

      // Stage 4: Verify JavaScript output has specific input functions
      expect(jsCode).toContain('input.int(');
      expect(jsCode).toContain('input.float(');
      expect(jsCode).toContain('input.bool(');

      // Stage 5: Verify type parameter was removed (not passed to specific functions)
      expect(jsCode).not.toContain('type:');
    }, TRANSPILER_TIMEOUT);

    it('should handle mixed input syntax in full pipeline', async () => {
      const v4Code = `//@version=4
indicator("Test")
val1 = input(10, type=input.integer)
val2 = input(1.5, type=input.float)`;

      const migratedCode = PineVersionMigrator.migrate(v4Code, 4);
      const jsCode = await transpiler.transpile(migratedCode);

      expect(jsCode).toBeDefined();
      expect(jsCode).toContain('input.int(');
      expect(jsCode).toContain('input.float(');
    }, TRANSPILER_TIMEOUT);
  });
});
