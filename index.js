import { PineTS, Provider } from '../PineTS/dist/pinets.dev.es.js';
import { writeFileSync } from 'fs';
import { MoexProvider } from './providers/MoexProvider.js';

/* Extended Provider System */
const PROVIDERS = {
    Binance: Provider.Binance,
    MOEX: new MoexProvider()
};

/* Symbol-to-Provider Auto-Detection */
const SYMBOL_MAPPING = {
    // MOEX Stocks
    'SBER': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: 'D' },
    'GAZP': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: '60' },
    'YNDX': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: 'D' },
    'LKOH': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: 'D' },
    'ROSN': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: 'D' },
    'NVTK': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: 'D' },
    'MGNT': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: 'D' },
    'VTBR': { provider: 'MOEX', exchange: 'MOEX', defaultTimeframe: 'D' },
    
    // Binance Crypto
    'BTCUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '1m' },
    'ETHUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '4h' },
    'ADAUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '1h' },
    'DOTUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '1h' },
    'LINKUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '1h' },
    'UNIUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '1h' },
    'LTCUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '1h' },
    'XRPUSDT': { provider: 'Binance', exchange: 'Binance', defaultTimeframe: '1h' }
};

/* Default Configuration - Auto-detects provider based on symbol */
const DEFAULT_CONFIG = {
    symbol: process.env.SYMBOL || 'BTCUSDT',
    timeframe: process.env.TIMEFRAME || null, // Auto-detect if not specified
    bars: parseInt(process.env.BARS) || 100,
    strategy: 'EMA Crossover Strategy',
    indicators: {
        ema9: { period: 9, color: '#2196F3' },
        ema18: { period: 18, color: '#F44336' },
        signals: { color: '#4CAF50' }
    }
};

/* Alternative configurations for easy switching */
const PRESET_CONFIGS = {
    binance_scalping: {
        symbol: 'BTCUSDT',
        provider: PROVIDERS.Binance,
        exchange: 'Binance',
        timeframe: '1m',
        bars: 200,
        strategy: 'Binance Scalping EMA',
        indicators: {
            ema9: { period: 9, color: '#00BCD4' },
            ema21: { period: 21, color: '#FF5722' }
        }
    },
    binance_swing: {
        symbol: 'ETHUSDT', 
        provider: PROVIDERS.Binance,
        exchange: 'Binance',
        timeframe: '4h',
        bars: 150,
        strategy: 'Binance Swing Trading',
        indicators: {
            ema20: { period: 20, color: '#9C27B0' },
            ema50: { period: 50, color: '#FF9800' }
        }
    },
    moex_daily: {
        symbol: 'SBER',
        provider: PROVIDERS.MOEX,
        exchange: 'MOEX',
        timeframe: 'D',
        bars: 100,
        strategy: 'MOEX Daily EMA',
        indicators: {
            ema9: { period: 9, color: '#2196F3' },
            ema18: { period: 18, color: '#F44336' }
        }
    },
    moex_intraday: {
        symbol: 'GAZP',
        provider: PROVIDERS.MOEX,
        exchange: 'MOEX', 
        timeframe: '60',
        bars: 150,
        strategy: 'MOEX Hourly Trading',
        indicators: {
            ema12: { period: 12, color: '#4CAF50' },
            ema26: { period: 26, color: '#FF5722' }
        }
    }
};

class ConfigurationManager {
    static autoDetectProvider(symbol) {
        const symbolInfo = SYMBOL_MAPPING[symbol.toUpperCase()];
        if (!symbolInfo) {
            throw new Error(`Unknown symbol: ${symbol}. Add it to SYMBOL_MAPPING or use supported symbols.`);
        }
        return symbolInfo;
    }

    static createUnifiedConfig(symbol, timeframe = null, bars = 100) {
        const symbolInfo = this.autoDetectProvider(symbol);
        
        return {
            symbol: symbol.toUpperCase(),
            provider: PROVIDERS[symbolInfo.provider],
            exchange: symbolInfo.exchange,
            timeframe: timeframe || symbolInfo.defaultTimeframe,
            bars,
            strategy: `${symbolInfo.exchange} Auto-Detected Strategy`,
            indicators: DEFAULT_CONFIG.indicators
        };
    }

    static generateChartConfig(tradingConfig, calculatedIndicators) {
        return {
            ui: {
                title: `${tradingConfig.strategy} - ${tradingConfig.symbol} (${tradingConfig.exchange})`,
                symbol: tradingConfig.symbol,
                exchange: tradingConfig.exchange,
                timeframe: this.formatTimeframe(tradingConfig.timeframe),
                strategy: tradingConfig.strategy
            },
            dataSource: {
                url: "chart-data.json",
                candlestickPath: "candlestick",
                plotsPath: "plots",
                timestampPath: "timestamp"
            },
            chartLayout: {
                main: { height: 400 },
                indicator: { height: 200 }
            },
            seriesConfig: {
                candlestick: {
                    upColor: "#26a69a",
                    downColor: "#ef5350",
                    borderVisible: false,
                    wickUpColor: "#26a69a",
                    wickDownColor: "#ef5350"
                },
                series: this.generateSeriesConfig(calculatedIndicators)
            }
        };
    }

    static formatTimeframe(timeframe) {
        const timeframes = {
            '1': '1 Minute',
            '5': '5 Minutes', 
            '10': '10 Minutes',
            '15': '15 Minutes',
            '30': '30 Minutes',
            '60': '1 Hour',
            '240': '4 Hours',
            'D': 'Daily',
            'W': 'Weekly',
            'M': 'Monthly'
        };
        return timeframes[timeframe] || timeframe;
    }

    static generateSeriesConfig(indicators) {
        const seriesConfig = {};
        const colors = ['#2196F3', '#F44336', '#4CAF50', '#FF9800', '#9C27B0', '#00BCD4'];
        let colorIndex = 0;

        Object.entries(indicators).forEach(([key, indicator]) => {
            if (key.includes('EMA') || key.includes('SMA') || key.includes('MA')) {
                seriesConfig[key] = {
                    color: colors[colorIndex % colors.length],
                    lineWidth: 2,
                    title: indicator.title || key,
                    chart: 'main'
                };
            } else {
                seriesConfig[key] = {
                    color: colors[colorIndex % colors.length],
                    lineWidth: 2,
                    title: indicator.title || key,
                    chart: 'indicator'
                };
            }
            colorIndex++;
        });

        return seriesConfig;
    }

    static exportConfiguration(config) {
        writeFileSync('chart-config.json', JSON.stringify(config, null, 2));
    }
}

class MarketDataProcessor {
    static validateOHLCV(candle) {
        const { open, high, low, close, volume } = candle;
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
            .filter(this.validateOHLCV)
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

class TechnicalAnalysisCalculator {
    static async runEMAIndicator(pineTS) {
        return await pineTS.run((context) => {
            const ta = context.ta;
            const math = context.math;
            const { close, open } = context.data;
            const { plot } = context.core;
            
            const ema9 = ta.ema(close, 9);
            const ema18 = ta.ema(close, 18);
            const bullBias = ema9 > ema18;
            
            plot(ema9, 'EMA9', { style: 'line', linewidth: 2, color: 'blue' });
            plot(ema18, 'EMA18', { style: 'line', linewidth: 2, color: 'red' });
            plot(bullBias ? 1 : 0, 'BullSignal', { style: 'line', linewidth: 1, color: 'green' });
            
            return {
                ema9,
                ema18,
                bullBias,
                prevClose: close[1],
                diffClose: close - close[1],
                absDiff: math.abs(open[1] - close[2])
            };
        });
    }

    static getIndicatorMetadata() {
        return {
            EMA9: { title: 'EMA 9', type: 'moving_average' },
            EMA18: { title: 'EMA 18', type: 'moving_average' },
            BullSignal: { title: 'Bull Signal', type: 'signal' }
        };
    }
}

class ChartDataExporter {
    static exportToFile(candlestickData, plots) {
        const chartData = {
            candlestick: candlestickData,
            plots,
            timestamp: new Date().toISOString()
        };
        
        writeFileSync('chart-data.json', JSON.stringify(chartData, null, 2));
    }
}

async function main() {
    try {
        // Unified configuration - auto-detects provider based on symbol
        const symbol = process.env.SYMBOL || DEFAULT_CONFIG.symbol;
        const timeframe = process.env.TIMEFRAME || DEFAULT_CONFIG.timeframe;
        const bars = parseInt(process.env.BARS) || DEFAULT_CONFIG.bars;
        
        // Create unified config with auto-detection
        const activeConfig = ConfigurationManager.createUnifiedConfig(symbol, timeframe, bars);
        
        console.log(`Auto-detected: Trading ${activeConfig.symbol} on ${activeConfig.exchange} (${activeConfig.timeframe})`);
        
        const pineTS = new PineTS(
            activeConfig.provider, 
            activeConfig.symbol, 
            activeConfig.timeframe, 
            activeConfig.bars
        );
        
        const { result, plots } = await TechnicalAnalysisCalculator.runEMAIndicator(pineTS);
        const indicatorMetadata = TechnicalAnalysisCalculator.getIndicatorMetadata();
        
        await pineTS.ready();
        const rawMarketData = pineTS.data;
        
        const candlestickData = rawMarketData?.length > 0 
            ? MarketDataProcessor.processCandlestickData(rawMarketData)
            : MarketDataProcessor.createSyntheticData(result.ema9, result.ema18);
        
        ChartDataExporter.exportToFile(candlestickData, plots);
        
        /* Generate and export chart configuration */
        const chartConfig = ConfigurationManager.generateChartConfig(activeConfig, indicatorMetadata);
        ConfigurationManager.exportConfiguration(chartConfig);
        
        console.log(`Successfully processed ${candlestickData.length} candles for ${activeConfig.symbol}`);
        
    } catch (error) {
        console.error('Error:', error);
        process.exit(1);
    }
}

main();