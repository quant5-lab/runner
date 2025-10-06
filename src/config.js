import { Provider } from '../../PineTS/dist/pinets.dev.es.js';
import { MoexProvider } from './providers/MoexProvider.js';
import { YahooFinanceProvider } from './providers/YahooFinanceProvider.js';

/* Provider chain factory - requires logger injection */
export function createProviderChain(logger) {
  return [
    { name: 'MOEX', instance: new MoexProvider(logger) },
    { name: 'Binance', instance: Provider.Binance },
    { name: 'YahooFinance', instance: new YahooFinanceProvider(logger) },
  ];
}

/* Default application configuration */
export const DEFAULTS = {
  symbol: process.env.SYMBOL || 'BTCUSDT',
  timeframe: process.env.TIMEFRAME || 'D',
  bars: parseInt(process.env.BARS) || 100,
  strategy: 'EMA Crossover Strategy',
};
