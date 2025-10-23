import { spawn } from 'child_process';
import { readFile, writeFile, unlink } from 'fs/promises';
import { createHash } from 'crypto';
import escodegen from 'escodegen';

class PineScriptTranspilationError extends Error {
  constructor(message, cause) {
    super(message);
    this.name = 'PineScriptTranspilationError';
    this.cause = cause;
  }
}

export class PineScriptTranspiler {
  constructor(logger) {
    this.logger = logger;
    this.cache = new Map();
  }

  async transpile(pineScriptCode) {
    const cacheKey = this.getCacheKey(pineScriptCode);

    if (this.cache.has(cacheKey)) {
      this.logger.info('Cache hit for Pine Script transpilation');
      return this.cache.get(cacheKey);
    }

    try {
      const version = this.detectVersion(pineScriptCode);
      const timestamp = Date.now();
      const inputPath = `/tmp/input-${timestamp}.pine`;
      const outputPath = `/tmp/output-${timestamp}.json`;

      await this.writeTempPineFile(inputPath, pineScriptCode);

      const ast = await this.spawnPythonParser(inputPath, outputPath, version);

      const jsCode = this.generateJavaScript(ast);

      await this.cleanupTempFiles(inputPath, outputPath);

      this.cache.set(cacheKey, jsCode);

      return jsCode;
    } catch (error) {
      throw new PineScriptTranspilationError(
        `Failed to transpile Pine Script: ${error.message}`,
        error,
      );
    }
  }

  async spawnPythonParser(inputPath, outputPath, version) {
    return new Promise((resolve, reject) => {
      const args = ['services/pine-parser/parser.py', inputPath, outputPath];
      const pythonProcess = spawn('python3', args);

      let stderr = '';

      pythonProcess.stderr.on('data', (data) => {
        stderr += data.toString();
      });

      pythonProcess.on('close', async (code) => {
        if (code !== 0) {
          reject(new Error(`Python parser exited with code ${code}: ${stderr}`));
          return;
        }

        try {
          const astJson = await this.readAstFromJson(outputPath);
          resolve(astJson);
        } catch (error) {
          reject(new Error(`Failed to read AST from ${outputPath}: ${error.message}`));
        }
      });

      pythonProcess.on('error', (error) => {
        reject(new Error(`Failed to spawn Python parser: ${error.message}`));
      });
    });
  }

  async writeTempPineFile(filePath, content) {
    await writeFile(filePath, content, 'utf-8');
  }

  async readAstFromJson(filePath) {
    const jsonContent = await readFile(filePath, 'utf-8');
    return JSON.parse(jsonContent);
  }

  async cleanupTempFiles(...filePaths) {
    for (const filePath of filePaths) {
      try {
        await unlink(filePath);
      } catch (error) {
        this.logger.warn(`Failed to cleanup temp file ${filePath}: ${error.message}`);
      }
    }
  }

  transformStrategyCall(node) {
    if (!node || typeof node !== 'object') return;

    /* Transform strategy() → strategy.call() for PineTS compatibility */
    if (
      node.type === 'CallExpression' &&
      node.callee &&
      node.callee.type === 'Identifier' &&
      node.callee.name === 'strategy'
    ) {
      node.callee = {
        type: 'MemberExpression',
        object: { type: 'Identifier', name: 'strategy' },
        property: { type: 'Identifier', name: 'call' },
        computed: false,
      };
    }

    /* Recursively process all node properties */
    for (const key in node) {
      if (Object.prototype.hasOwnProperty.call(node, key) && key !== 'loc' && key !== 'range') {
        const value = node[key];
        if (Array.isArray(value)) {
          for (let i = 0; i < value.length; i++) {
            this.transformStrategyCall(value[i]);
          }
        } else if (typeof value === 'object' && value !== null) {
          this.transformStrategyCall(value);
        }
      }
    }
  }

  wrapHistoricalReferences(node) {
    if (!node || typeof node !== 'object') return;

    // Wrap MemberExpression with historical index (e.g., counter[1] -> (counter[1] || 0))
    if (
      node.type === 'MemberExpression' &&
      node.computed &&
      node.property &&
      node.property.type === 'Literal' &&
      node.property.value > 0
    ) {
      // Return wrapped node
      return {
        type: 'LogicalExpression',
        operator: '||',
        left: node,
        right: { type: 'Literal', value: 0, raw: '0' },
      };
    }

    // Recursively process all node properties
    for (const key in node) {
      if (Object.prototype.hasOwnProperty.call(node, key) && key !== 'loc' && key !== 'range') {
        const value = node[key];
        if (Array.isArray(value)) {
          for (let i = 0; i < value.length; i++) {
            const wrapped = this.wrapHistoricalReferences(value[i]);
            if (wrapped && wrapped !== value[i]) {
              value[i] = wrapped;
            }
          }
        } else if (typeof value === 'object' && value !== null) {
          const wrapped = this.wrapHistoricalReferences(value);
          if (wrapped && wrapped !== value) {
            node[key] = wrapped;
          }
        }
      }
    }

    return node;
  }

  generateJavaScript(ast) {
    try {
      // Transform strategy() → strategy.call() for PineTS compatibility
      this.transformStrategyCall(ast);

      // Transform AST to wrap historical references with || 0
      this.wrapHistoricalReferences(ast);

      return escodegen.generate(ast, {
        format: {
          indent: {
            style: '  ',
          },
          quotes: 'single',
        },
      });
    } catch (error) {
      throw new Error(`escodegen failed: ${error.message}`);
    }
  }

  detectVersion(pineScriptCode) {
    const firstLine = pineScriptCode.split('\n')[0];
    const versionMatch = firstLine.match(/\/\/@version=(\d+)/);

    if (versionMatch) {
      return parseInt(versionMatch[1]);
    }

    return 5;
  }

  getCacheKey(pineScriptCode) {
    return createHash('sha256').update(pineScriptCode).digest('hex');
  }
}

export { PineScriptTranspilationError };
