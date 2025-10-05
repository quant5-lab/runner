import { ProviderManager } from './classes/ProviderManager.js';
import { TechnicalAnalysisEngine } from './classes/TechnicalAnalysisEngine.js';
import { DataProcessor } from './classes/DataProcessor.js';
import { ConfigurationBuilder } from './classes/ConfigurationBuilder.js';
import { FileExporter } from './classes/FileExporter.js';
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

  container
    .register('logger', () => new Logger(), true)
    .register('providerManager', () => new ProviderManager(providerChain), true)
    .register('technicalAnalysisEngine', () => new TechnicalAnalysisEngine(), true)
    .register('dataProcessor', () => new DataProcessor(), true)
    .register('configurationBuilder', () => new ConfigurationBuilder(defaults), true)
    .register('fileExporter', () => new FileExporter(), true)
    .register(
      'tradingOrchestrator',
      (container) =>
        new TradingOrchestrator(
          container.resolve('providerManager'),
          container.resolve('technicalAnalysisEngine'),
          container.resolve('dataProcessor'),
          container.resolve('configurationBuilder'),
          container.resolve('fileExporter'),
          container.resolve('logger'),
        ),
      true,
    );

  return container;
}

export { Container, createContainer };
