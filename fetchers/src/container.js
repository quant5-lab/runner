import { ProviderManager } from './classes/ProviderManager.js';
import { Logger } from './classes/Logger.js';
import ApiStatsCollector from './utils/ApiStatsCollector.js';

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
    .register('apiStatsCollector', () => new ApiStatsCollector(), true)
    .register(
      'providerManager',
      (c) =>
        new ProviderManager(
          providerChain(logger, c.resolve('apiStatsCollector')),
          c.resolve('logger'),
        ),
      true,
    );

  return container;
}

export { Container, createContainer };
