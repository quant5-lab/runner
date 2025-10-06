import { MoexProvider } from './providers/MoexProvider.js';
import { BinanceProvider } from './providers/BinanceProvider.js';
import { YahooFinanceProvider } from './providers/YahooFinanceProvider.js';

/* Provider chain factory - requires logger injection */
export function createProviderChain(logger) {
  return [
    { name: 'MOEX', instance: new MoexProvider(logger) },
    { name: 'Binance', instance: new BinanceProvider(logger) },
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
