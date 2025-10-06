import { describe, it, expect, beforeEach, vi } from 'vitest';
import { JsonFileWriter } from '../classes/JsonFileWriter.js';
import * as fs from 'fs';

vi.mock('fs', () => ({
  writeFileSync: vi.fn(),
  mkdirSync: vi.fn(),
}));

vi.mock('path', () => ({
  join: vi.fn((...args) => args.join('/')),
}));

describe('JsonFileWriter', () => {
  let exporter;

  beforeEach(() => {
    exporter = new JsonFileWriter();
    vi.clearAllMocks();
  });

  describe('ensureOutDirectory()', () => {
    it('should create out directory', () => {
      exporter.ensureOutDirectory();
      expect(fs.mkdirSync).toHaveBeenCalledWith('out', { recursive: true });
    });

    it('should handle existing directory error', () => {
      fs.mkdirSync.mockImplementationOnce(() => {
        throw new Error('EEXIST');
      });
      expect(() => exporter.ensureOutDirectory()).not.toThrow();
    });
  });

  describe('exportChartData()', () => {
    it('should export chart data with candlestick and plots', () => {
      const candlestickData = [
        { time: 1, open: 100, high: 105, low: 95, close: 102 },
        { time: 2, open: 102, high: 108, low: 100, close: 107 },
      ];
      const plots = {
        EMA20: [100, 101, 102],
        RSI: [45, 50, 55],
      };

      exporter.exportChartData(candlestickData, plots);

      expect(fs.mkdirSync).toHaveBeenCalledWith('out', { recursive: true });
      expect(fs.writeFileSync).toHaveBeenCalledTimes(1);

      const writeCall = fs.writeFileSync.mock.calls[0];
      expect(writeCall[0]).toBe('out/chart-data.json');

      const written = JSON.parse(writeCall[1]);
      expect(written.candlestick).toEqual(candlestickData);
      expect(written.plots).toEqual(plots);
      expect(written.timestamp).toBeDefined();
      expect(typeof written.timestamp).toBe('string');
    });

    it('should include ISO timestamp', () => {
      const candlestickData = [];
      const plots = {};

      exporter.exportChartData(candlestickData, plots);

      const writeCall = fs.writeFileSync.mock.calls[0];
      const written = JSON.parse(writeCall[1]);
      const timestamp = new Date(written.timestamp);
      expect(timestamp.toISOString()).toBe(written.timestamp);
    });

    it('should handle empty data', () => {
      exporter.exportChartData([], {});

      expect(fs.writeFileSync).toHaveBeenCalled();
      const writeCall = fs.writeFileSync.mock.calls[0];
      const written = JSON.parse(writeCall[1]);
      expect(written.candlestick).toEqual([]);
      expect(written.plots).toEqual({});
    });

    it('should format JSON with 2 space indentation', () => {
      exporter.exportChartData([{ time: 1 }], {});

      const writeCall = fs.writeFileSync.mock.calls[0];
      expect(writeCall[1]).toContain('\n  ');
    });
  });

  describe('exportConfiguration()', () => {
    it('should export configuration to file', () => {
      const config = {
        ui: { title: 'Test Chart' },
        dataSource: { url: 'chart-data.json' },
        chartLayout: { main: { height: 400 } },
      };

      exporter.exportConfiguration(config);

      expect(fs.mkdirSync).toHaveBeenCalledWith('out', { recursive: true });
      expect(fs.writeFileSync).toHaveBeenCalledTimes(1);

      const writeCall = fs.writeFileSync.mock.calls[0];
      expect(writeCall[0]).toBe('out/chart-config.json');

      const written = JSON.parse(writeCall[1]);
      expect(written).toEqual(config);
    });

    it('should format JSON with 2 space indentation', () => {
      const config = { test: 'value' };
      exporter.exportConfiguration(config);

      const writeCall = fs.writeFileSync.mock.calls[0];
      expect(writeCall[1]).toContain('\n  ');
    });

    it('should handle complex nested configuration', () => {
      const config = {
        ui: { title: 'Complex', nested: { deep: { value: 123 } } },
        arrays: [1, 2, 3],
        objects: { a: { b: { c: 'd' } } },
      };

      exporter.exportConfiguration(config);

      const writeCall = fs.writeFileSync.mock.calls[0];
      const written = JSON.parse(writeCall[1]);
      expect(written).toEqual(config);
    });

    it('should handle empty configuration', () => {
      exporter.exportConfiguration({});

      expect(fs.writeFileSync).toHaveBeenCalled();
      const writeCall = fs.writeFileSync.mock.calls[0];
      expect(JSON.parse(writeCall[1])).toEqual({});
    });
  });
});
