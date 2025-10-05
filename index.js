import { PineTS, Provider } from '../PineTS/dist/pinets.dev.es.js';
import { writeFileSync } from 'fs';

/* Configuration - Single Source of Truth */
const TRADING_CONFIG = {
    symbol: 'BTCUSDT',
    provider: Provider.Binance,
    timeframe: 'D',
    bars: 100,
    strategy: 'EMA Crossover Strategy',
    indicators: {
        ema9: { period: 9, color: '#2196F3' },
        ema18: { period: 18, color: '#F44336' },
        signals: { color: '#4CAF50' }
    }
};

/* Alternative configurations for easy switching */
const PRESET_CONFIGS = {
    scalping: {
        symbol: 'BTCUSDT',
        timeframe: '1m',
        bars: 200,
        strategy: 'Scalping EMA',
        indicators: {
            ema9: { period: 9, color: '#00BCD4' },
            ema21: { period: 21, color: '#FF5722' }
        }
    },
    swing: {
        symbol: 'ETHUSDT', 
        timeframe: '4h',
        bars: 150,
        strategy: 'Swing Trading',
        indicators: {
            ema20: { period: 20, color: '#9C27B0' },
            ema50: { period: 50, color: '#FF9800' }
        }
    }
};

class ConfigurationManager {
    static generateChartConfig(tradingConfig, calculatedIndicators) {
        return {
            ui: {
                title: `${tradingConfig.strategy} - ${tradingConfig.symbol}`,
                symbol: tradingConfig.symbol,
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
            '1m': '1 Minute',
            '5m': '5 Minutes', 
            '15m': '15 Minutes',
            '1h': '1 Hour',
            '4h': '4 Hours',
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
        const pineTS = new PineTS(
            TRADING_CONFIG.provider, 
            TRADING_CONFIG.symbol, 
            TRADING_CONFIG.timeframe, 
            TRADING_CONFIG.bars
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
        const chartConfig = ConfigurationManager.generateChartConfig(TRADING_CONFIG, indicatorMetadata);
        ConfigurationManager.exportConfiguration(chartConfig);
        
    } catch (error) {
        console.error('Error:', error);
        process.exit(1);
    }
}

main();