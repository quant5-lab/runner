import { ProviderManager } from './classes/ProviderManager.js';
import { TechnicalAnalysisEngine } from './classes/TechnicalAnalysisEngine.js';
import { DataProcessor } from './classes/DataProcessor.js';
import { ConfigurationBuilder } from './classes/ConfigurationBuilder.js';
import { JsonFileWriter } from './classes/JsonFileWriter.js';
import { TradingOrchestrator } from './classes/TradingOrchestrator.js';
import { Logger } from './classes/Logger.js';

class Container {
  constructor() {
    this.services = new Map();
    this.singletons = new Map();
  }

  register(name, factory, singleton = false) {
    this.services.set(name, { factory, singleton });
    return this;
  }

  resolve(name) {
    const service = this.services.get(name);
    if (!service) {
      throw new Error(`Service ${name} not registered`);
    }

    if (service.singleton) {
      if (!this.singletons.has(name)) {
        this.singletons.set(name, service.factory(this));
      }
      return this.singletons.get(name);
    }

    return service.factory(this);
  }
}

function createContainer(providerChain, defaults) {
  const container = new Container();
  const logger = new Logger();

  container
    .register('logger', () => logger, true)
    .register(
      'providerManager',
      (c) => new ProviderManager(providerChain(logger), c.resolve('logger')),
      true,
    )
    .register('technicalAnalysisEngine', () => new TechnicalAnalysisEngine(), true)
    .register('dataProcessor', () => new DataProcessor(), true)
    .register('configurationBuilder', (c) => new ConfigurationBuilder(defaults), true)
    .register('jsonFileWriter', () => new JsonFileWriter(), true)
    .register(
      'tradingOrchestrator',
      (c) =>
        new TradingOrchestrator(
          c.resolve('providerManager'),
          c.resolve('technicalAnalysisEngine'),
          c.resolve('dataProcessor'),
          c.resolve('configurationBuilder'),
          c.resolve('jsonFileWriter'),
          c.resolve('logger'),
        ),
      true,
    );

  return container;
}

export { Container, createContainer };
