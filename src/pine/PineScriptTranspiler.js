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

      pythonProcess.on('close', async(code) => {
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

  generateJavaScript(ast) {
    try {
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
      const version = parseInt(versionMatch[1]);
      if (version === 4 || version === 5) {
        return version;
      }
      this.logger.warn(`Unsupported Pine Script version: ${version}, defaulting to 5`);
      return 5;
    }

    this.logger.warn('No //@version comment found, defaulting to version 5');
    return 5;
  }

  getCacheKey(pineScriptCode) {
    return createHash('sha256').update(pineScriptCode).digest('hex');
  }
}

export { PineScriptTranspilationError };
