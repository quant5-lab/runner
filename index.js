import { PineTS, Provider } from '../PineTS/dist/pinets.dev.es.js';
import { writeFileSync, mkdirSync } from 'fs';
import { join } from 'path';
import { MoexProvider } from './providers/MoexProvider.js';
import { YahooFinanceProvider } from './providers/YahooFinanceProvider.js';

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

class ProviderManager {
    static async fetchMarketData(symbol, timeframe, bars) {
        for (let i = 0; i < PROVIDER_CHAIN.length; i++) {
            const { name, instance } = PROVIDER_CHAIN[i];
            
            try {
                const marketData = await this.executeProvider(name, instance, symbol, timeframe, bars);
                
                if (marketData?.length > 0) {
                    return { provider: name, data: marketData, instance };
                }
            } catch (error) {
                continue;
            }
        }
        
        throw new Error(`All providers failed for symbol: ${symbol}`);
    }

    static async executeProvider(name, instance, symbol, timeframe, bars) {
        if (name === 'Binance') {
            const pineTS = new PineTS(instance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS.data;
        }
        
        return await instance.getMarketData(symbol, timeframe, bars);
    }
}

class TechnicalAnalysisEngine {
    static async createPineTSAdapter(provider, data, instance, symbol, timeframe, bars) {
        if (provider === 'Binance') {
            // Binance: Use PineTS built-in Binance provider - exact pattern from docs
            const pineTS = new PineTS(Provider.Binance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        } else if (provider === 'MOEX') {
            // MOEX: Use existing MOEX provider instance
            const pineTS = new PineTS(instance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        } else if (provider === 'YahooFinance') {
            // Yahoo: Use existing Yahoo provider instance
            const pineTS = new PineTS(instance, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        } else {
            // Fallback: Pass data array directly
            const pineTS = new PineTS(data, symbol, timeframe, bars);
            await pineTS.ready();
            return pineTS;
        }
    }

    static async runEMAStrategy(pineTS) {
        /* Add signal calculation using proper PineTS syntax */
        const { plots } = await pineTS.run((context) => {
            const { close } = context.data;
            const { plot } = context.core;
            const ta = context.ta;
            
            // Basic EMA calculation
            const ema9 = ta.ema(close, 9);
            const ema18 = ta.ema(close, 18);
            
            // Bull signal using Pine Script style comparison - current values
            const bullSignal = ema9 > ema18 ? 1 : 0;
            
            // Plot calls - exact pattern from docs
            plot(ema9, 'EMA9', { style: 'line', linewidth: 2, color: 'blue' });
            plot(ema18, 'EMA18', { style: 'line', linewidth: 2, color: 'red' });
            plot(bullSignal, 'BullSignal', { style: 'line', linewidth: 1, color: 'green' });
        });
        
        return { result: plots, plots: plots || {} };
    }

    static getIndicatorMetadata() {
        return {
            EMA9: { title: 'EMA 9', type: 'moving_average' },
            EMA18: { title: 'EMA 18', type: 'moving_average' },
            BullSignal: { title: 'Bull Signal', type: 'signal' }
        };
    }
}

class DataProcessor {
    static isValidCandle(candle) {
        const { open, high, low, close } = candle;
        const values = [open, high, low, close].map(parseFloat);
        
        return values.every(val => !isNaN(val) && val > 0) && 
               Math.max(...values) === parseFloat(high) &&
               Math.min(...values) === parseFloat(low);
    }

    static normalizeCandle(candle) {
        const open = parseFloat(candle.open);
        const high = parseFloat(candle.high);
        const low = parseFloat(candle.low);
        const close = parseFloat(candle.close);
        const volume = parseFloat(candle.volume) || 1000;
        
        return {
            time: Math.floor(candle.openTime / 1000),
            open,
            high: Math.max(open, high, low, close),
            low: Math.min(open, high, low, close),
            close,
            volume
        };
    }

    static processCandlestickData(rawData) {
        if (!rawData?.length) return [];
        
        return rawData
            .filter(this.isValidCandle)
            .map(this.normalizeCandle);
    }

    static createSyntheticData(ema9, ema18) {
        const now = Date.now();
        return ema9.map((closeVal, i) => {
            const openVal = ema18[i];
            return {
                time: Math.floor((now - (ema9.length - i - 1) * 86400000) / 1000),
                open: openVal,
                high: Math.max(openVal, closeVal) * 1.002,
                low: Math.min(openVal, closeVal) * 0.998,
                close: closeVal,
                volume: 1000
            };
        });
    }
}

class ConfigurationBuilder {
    static createTradingConfig(symbol, timeframe = 'D', bars = 100) {
        return {
            symbol: symbol.toUpperCase(),
            timeframe,
            bars,
            strategy: 'Multi-Provider Strategy',
            indicators: DEFAULT_CONFIG.indicators
        };
    }

    static generateChartConfig(tradingConfig, indicatorMetadata) {
        return {
            ui: this.buildUIConfig(tradingConfig),
            dataSource: this.buildDataSourceConfig(),
            chartLayout: this.buildLayoutConfig(),
            seriesConfig: {
                candlestick: {
                    upColor: "#26a69a",
                    downColor: "#ef5350",
                    borderVisible: false,
                    wickUpColor: "#26a69a",
                    wickDownColor: "#ef5350"
                },
                series: this.buildSeriesConfig(indicatorMetadata)
            }
        };
    }

    static buildUIConfig(tradingConfig) {
        return {
            title: `${tradingConfig.strategy} - ${tradingConfig.symbol}`,
            symbol: tradingConfig.symbol,
            timeframe: this.formatTimeframe(tradingConfig.timeframe),
            strategy: tradingConfig.strategy
        };
    }

    static buildDataSourceConfig() {
        return {
            url: "chart-data.json",
            candlestickPath: "candlestick",
            plotsPath: "plots",
            timestampPath: "timestamp"
        };
    }

    static buildLayoutConfig() {
        return {
            main: { height: 400 },
            indicator: { height: 200 }
        };
    }

    static buildSeriesConfig(indicators) {
        const seriesConfig = {};
        const colors = ['#2196F3', '#F44336', '#4CAF50', '#FF9800', '#9C27B0', '#00BCD4'];
        let colorIndex = 0;

        Object.entries(indicators).forEach(([key, indicator]) => {
            seriesConfig[key] = {
                color: colors[colorIndex % colors.length],
                lineWidth: 2,
                title: indicator.title || key,
                chart: this.determineChartType(key)
            };
            colorIndex++;
        });

        return seriesConfig;
    }

    static determineChartType(key) {
        return (key.includes('EMA') || key.includes('SMA') || key.includes('MA')) ? 'main' : 'indicator';
    }

    static formatTimeframe(timeframe) {
        const timeframes = {
            '1': '1 Minute', '5': '5 Minutes', '10': '10 Minutes',
            '15': '15 Minutes', '30': '30 Minutes', '60': '1 Hour',
            '240': '4 Hours', 'D': 'Daily', 'W': 'Weekly', 'M': 'Monthly'
        };
        return timeframes[timeframe] || timeframe;
    }
}

class FileExporter {
    static ensureOutDirectory() {
        try {
            mkdirSync('out', { recursive: true });
        } catch (error) {
            /* Directory already exists */
        }
    }

    static exportChartData(candlestickData, plots) {
        this.ensureOutDirectory();
        const chartData = {
            candlestick: candlestickData,
            plots,
            timestamp: new Date().toISOString()
        };
        
        writeFileSync(join('out', 'chart-data.json'), JSON.stringify(chartData, null, 2));
    }

    static exportConfiguration(config) {
        this.ensureOutDirectory();
        writeFileSync(join('out', 'chart-config.json'), JSON.stringify(config, null, 2));
    }
}

async function main() {
    try {
        const { symbol, timeframe, bars } = DEFAULT_CONFIG;
        const envSymbol = process.env.SYMBOL || symbol;
        const envTimeframe = process.env.TIMEFRAME || timeframe;
        const envBars = parseInt(process.env.BARS) || bars;
        
        console.log(`📊 Configuration: Symbol=${envSymbol}, Timeframe=${envTimeframe}, Bars=${envBars}`);
        
        const tradingConfig = ConfigurationBuilder.createTradingConfig(envSymbol, envTimeframe, envBars);
        
        console.log(`🎯 Attempting to fetch ${envSymbol} (${envTimeframe}) with dynamic provider fallback`);
        
        const { provider, data, instance } = await ProviderManager.fetchMarketData(envSymbol, envTimeframe, envBars);
        
        console.log(`📊 Using ${provider} provider for ${envSymbol}`);
        
        const pineTS = await TechnicalAnalysisEngine.createPineTSAdapter(provider, data, instance, envSymbol, envTimeframe, envBars);
        
        const { result, plots } = await TechnicalAnalysisEngine.runEMAStrategy(pineTS);
        const indicatorMetadata = TechnicalAnalysisEngine.getIndicatorMetadata();
        
        // Process indicator plots - handle both custom providers and real PineTS
        let processedPlots = plots || {};
        
        if (result && Object.keys(processedPlots).length === 0) {
            // Ensure consistent time base for all indicators
            const createTimestamp = (i, length) => {
                return data[i]?.openTime 
                    ? Math.floor(data[i].openTime / 1000)
                    : Math.floor((Date.now() - (length - i - 1) * 86400000) / 1000);
            };
            
            // For real PineTS, extract plot data from result
            if (result.ema9 && Array.isArray(result.ema9)) {
                processedPlots.EMA9 = {
                    data: result.ema9.map((value, i) => ({
                        time: createTimestamp(i, result.ema9.length),
                        value: typeof value === 'number' ? value : 0
                    }))
                };
            }
            
            if (result.ema18 && Array.isArray(result.ema18)) {
                processedPlots.EMA18 = {
                    data: result.ema18.map((value, i) => ({
                        time: createTimestamp(i, result.ema18.length),
                        value: typeof value === 'number' ? value : 0
                    }))
                };
            }
            
            if (result.bullSignal !== undefined) {
                // Handle both single value and array
                const bullValues = Array.isArray(result.bullSignal) 
                    ? result.bullSignal 
                    : data.map((_, i) => i === data.length - 1 ? (result.bullSignal ? 1 : 0) : 0);
                    
                processedPlots.BullSignal = {
                    data: bullValues.map((value, i) => ({
                        time: createTimestamp(i, bullValues.length),
                        value: typeof value === 'boolean' ? (value ? 1 : 0) : (typeof value === 'number' ? value : 0)
                    }))
                };
            }
        }
        
        const candlestickData = data?.length > 0 
            ? DataProcessor.processCandlestickData(data)
            : DataProcessor.createSyntheticData(result.ema9, result.ema18);
        
        FileExporter.exportChartData(candlestickData, processedPlots);
        
        const chartConfig = ConfigurationBuilder.generateChartConfig(tradingConfig, indicatorMetadata);
        FileExporter.exportConfiguration(chartConfig);
        
        console.log(`Successfully processed ${candlestickData.length} candles for ${tradingConfig.symbol}`);
        
    } catch (error) {
        console.error('Error:', error);
        process.exit(1);
    }
}

main();