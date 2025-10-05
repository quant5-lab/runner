import { Provider } from '../PineTS/dist/pinets.dev.es.js';
import { MoexProvider } from './providers/MoexProvider.js';
import { YahooFinanceProvider } from './providers/YahooFinanceProvider.js';

/* Provider chain configuration */
export const PROVIDER_CHAIN = [
  { name: 'MOEX', instance: new MoexProvider() },
  { name: 'Binance', instance: Provider.Binance },
  { name: 'YahooFinance', instance: new YahooFinanceProvider() },
];

/* Default application configuration */
export const DEFAULTS = {
  symbol: process.env.SYMBOL || 'BTCUSDT',
  timeframe: process.env.TIMEFRAME || 'D',
  bars: parseInt(process.env.BARS) || 100,
  strategy: 'EMA Crossover Strategy',
  indicators: {
    ema9: { period: 9, color: '#2196F3' },
    ema18: { period: 18, color: '#F44336' },
    signals: { color: '#4CAF50' },
  },
};
