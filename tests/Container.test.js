import { describe, it, expect, beforeEach } from 'vitest';
import { Container, createContainer } from '../container.js';

describe('Container', () => {
  let container;

  beforeEach(() => {
    container = new Container();
  });

  describe('register()', () => {
    it('should register a service with factory', () => {
      const factory = () => ({ name: 'TestService' });
      container.register('test', factory);
      expect(container.services.has('test')).toBe(true);
    });

    it('should register singleton service', () => {
      const factory = () => ({ name: 'SingletonService' });
      container.register('singleton', factory, true);
      const service = container.services.get('singleton');
      expect(service.singleton).toBe(true);
    });

    it('should register non-singleton service', () => {
      const factory = () => ({ name: 'TransientService' });
      container.register('transient', factory, false);
      const service = container.services.get('transient');
      expect(service.singleton).toBe(false);
    });

    it('should return container for chaining', () => {
      const result = container.register('test', () => ({}));
      expect(result).toBe(container);
    });
  });

  describe('resolve()', () => {
    it('should resolve registered service', () => {
      const testService = { name: 'Test' };
      container.register('test', () => testService);
      const resolved = container.resolve('test');
      expect(resolved).toEqual(testService);
    });

    it('should throw error for unregistered service', () => {
      expect(() => container.resolve('nonexistent')).toThrow('Service nonexistent not registered');
    });

    it('should return same instance for singleton services', () => {
      let counter = 0;
      container.register('singleton', () => ({ id: ++counter }), true);
      const instance1 = container.resolve('singleton');
      const instance2 = container.resolve('singleton');
      expect(instance1).toBe(instance2);
      expect(instance1.id).toBe(1);
    });

    it('should return new instance for non-singleton services', () => {
      let counter = 0;
      container.register('transient', () => ({ id: ++counter }), false);
      const instance1 = container.resolve('transient');
      const instance2 = container.resolve('transient');
      expect(instance1).not.toBe(instance2);
      expect(instance1.id).toBe(1);
      expect(instance2.id).toBe(2);
    });

    it('should pass container to factory function', () => {
      let receivedContainer;
      container.register('test', (c) => {
        receivedContainer = c;
        return {};
      });
      container.resolve('test');
      expect(receivedContainer).toBe(container);
    });
  });
});

describe('createContainer', () => {
  it('should create container with all services registered', () => {
    const providerChain = ['MOEX', 'BINANCE'];
    const defaults = { SYMBOL: 'BTCUSDT', TIMEFRAME: 'D', BARS: 100 };
    const container = createContainer(providerChain, defaults);

    expect(container.services.has('logger')).toBe(true);
    expect(container.services.has('providerManager')).toBe(true);
    expect(container.services.has('technicalAnalysisEngine')).toBe(true);
    expect(container.services.has('dataProcessor')).toBe(true);
    expect(container.services.has('configurationBuilder')).toBe(true);
    expect(container.services.has('fileExporter')).toBe(true);
    expect(container.services.has('tradingOrchestrator')).toBe(true);
  });

  it('should register all services as singletons', () => {
    const container = createContainer([], {});
    const serviceNames = [
      'logger',
      'providerManager',
      'technicalAnalysisEngine',
      'dataProcessor',
      'configurationBuilder',
      'fileExporter',
      'tradingOrchestrator',
    ];

    serviceNames.forEach((name) => {
      const service = container.services.get(name);
      expect(service.singleton).toBe(true);
    });
  });

  it('should resolve logger instance', () => {
    const container = createContainer([], {});
    const logger = container.resolve('logger');
    expect(logger).toBeDefined();
    expect(typeof logger.log).toBe('function');
    expect(typeof logger.error).toBe('function');
  });

  it('should resolve providerManager with correct providerChain', () => {
    const mockProviderChain = (logger) => ['MOEX', 'YAHOO'];
    const container = createContainer(mockProviderChain, {});
    const providerManager = container.resolve('providerManager');
    expect(providerManager).toBeDefined();
    expect(providerManager.providerChain).toEqual(['MOEX', 'YAHOO']);
  });

  it('should resolve configurationBuilder with defaults', () => {
    const defaults = { SYMBOL: 'AAPL', TIMEFRAME: 'W' };
    const container = createContainer([], defaults);
    const configBuilder = container.resolve('configurationBuilder');
    expect(configBuilder).toBeDefined();
    expect(configBuilder.defaultConfig).toEqual(defaults);
  });

  it('should resolve tradingOrchestrator with all dependencies', () => {
    const mockProviderChain = (logger) => [];
    const container = createContainer(mockProviderChain, {});
    const orchestrator = container.resolve('tradingOrchestrator');
    expect(orchestrator).toBeDefined();
    expect(orchestrator.providerManager).toBeDefined();
    expect(orchestrator.technicalAnalysisEngine).toBeDefined();
    expect(orchestrator.dataProcessor).toBeDefined();
    expect(orchestrator.configurationBuilder).toBeDefined();
    expect(orchestrator.fileExporter).toBeDefined();
    expect(orchestrator.logger).toBeDefined();
  });
});
