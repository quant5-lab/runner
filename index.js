import { PineTS, Provider } from '../PineTS/dist/pinets.dev.es.js';
import { MoexProvider } from './providers/MoexProvider.js';
import { YahooFinanceProvider } from './providers/YahooFinanceProvider.js';
import { createContainer } from './container.js';

const PROVIDER_CHAIN = [
    { name: 'MOEX', instance: new MoexProvider() },
    { name: 'Binance', instance: Provider.Binance },
    { name: 'YahooFinance', instance: new YahooFinanceProvider() }
];

const DEFAULT_CONFIG = {
    symbol: process.env.SYMBOL || 'BTCUSDT',
    timeframe: process.env.TIMEFRAME || 'D',
    bars: parseInt(process.env.BARS) || 100,
    strategy: 'EMA Crossover Strategy',
    indicators: {
        ema9: { period: 9, color: '#2196F3' },
        ema18: { period: 18, color: '#F44336' },
        signals: { color: '#4CAF50' }
    }
};

async function main() {
    try {
        const { symbol, timeframe, bars } = DEFAULT_CONFIG;
        const envSymbol = process.env.SYMBOL || symbol;
        const envTimeframe = process.env.TIMEFRAME || timeframe;
        const envBars = parseInt(process.env.BARS) || bars;
        
        const container = createContainer(PROVIDER_CHAIN, DEFAULT_CONFIG);
        const orchestrator = container.resolve('tradingOrchestrator');
        const logger = container.resolve('logger');
        
        await orchestrator.runTradingAnalysis(envSymbol, envTimeframe, envBars);
        
    } catch (error) {
        const container = createContainer(PROVIDER_CHAIN, DEFAULT_CONFIG);
        const logger = container.resolve('logger');
        logger.error('Error:', error);
        process.exit(1);
    }
}

main();