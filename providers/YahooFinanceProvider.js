// Yahoo Finance Provider for PineTS - Real market data for US stocks
export class YahooFinanceProvider {
    constructor() {
        this.baseUrl = 'https://query1.finance.yahoo.com/v8/finance/chart';
        this.cache = new Map();
        this.cacheDuration = 5 * 60 * 1000; // 5 minutes
        this.headers = {
            'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36'
        };
    }

    /* Yahoo Finance interval mapping */
    static intervalMap = {
        '1': '1m',      // 1 minute
        '5': '5m',      // 5 minutes
        '15': '15m',    // 15 minutes
        '30': '30m',    // 30 minutes
        '60': '1h',     // 1 hour
        '240': '4h',    // 4 hours (not supported by Yahoo, use 1h)
        'D': '1d',      // Daily
        'W': '1wk',     // Weekly
        'M': '1mo'      // Monthly
    };

    /* Convert PineTS timeframe to Yahoo interval */
    convertTimeframe(timeframe) {
        return YahooFinanceProvider.intervalMap[timeframe] || '1d';
    }

    /* Generate cache key */
    getCacheKey(symbol, timeframe, range) {
        return `${symbol}_${timeframe}_${range}`;
    }

    /* Check cache */
    getFromCache(key) {
        const cached = this.cache.get(key);
        if (!cached) return null;
        
        if (Date.now() - cached.timestamp > this.cacheDuration) {
            this.cache.delete(key);
            return null;
        }
        
        return cached.data;
    }

    /* Set cache */
    setCache(key, data) {
        this.cache.set(key, {
            data,
            timestamp: Date.now()
        });
    }

    /* Convert Yahoo Finance data to PineTS format */
    convertYahooCandle(timestamp, open, high, low, close, volume) {
        return {
            openTime: timestamp * 1000, // Convert to milliseconds
            open: parseFloat(open),
            high: parseFloat(high),
            low: parseFloat(low),
            close: parseFloat(close),
            volume: parseFloat(volume || 0),
            closeTime: timestamp * 1000 + 60000, // Add 1 minute
            quoteAssetVolume: parseFloat(close) * parseFloat(volume || 0),
            numberOfTrades: 0,
            takerBuyBaseAssetVolume: 0,
            takerBuyQuoteAssetVolume: 0,
            ignore: null
        };
    }

    /* Get date range string for Yahoo API */
    getDateRange(limit, timeframe) {
        // Use ranges that provide sufficient data points (targeting ~100 like other providers)
        const ranges = {
            '1': '1d',        // 1 minute - 1 day gives ~390 points
            '5': '5d',        // 5 minutes - 5 days gives ~1440 points  
            '15': '5d',       // 15 minutes - 5 days gives ~480 points
            '30': '5d',       // 30 minutes - 5 days gives ~240 points
            '60': '5d',       // 1 hour - 5 days gives ~120 points
            '240': '1mo',     // 4 hours - 1 month gives ~180 points
            'D': '6mo',       // Daily - 6 months gives ~130 trading days
            'W': '2y',        // Weekly - 2 years gives ~104 weeks
            'M': '10y'        // Monthly - 10 years gives ~120 months
        };
        
        return ranges[timeframe] || '6mo';
    }

    /* Build Yahoo Finance API URL */
    buildUrl(symbol, interval, range) {
        const params = new URLSearchParams({
            interval: interval,
            range: range
        });
        
        return `${this.baseUrl}/${symbol}?${params.toString()}`;
    }

    /* Main method - get market data */
    async getMarketData(symbol, timeframe, limit = 100, sDate, eDate) {
        try {
            const interval = this.convertTimeframe(timeframe);
            const range = this.getDateRange(limit, timeframe);
            const cacheKey = this.getCacheKey(symbol, interval, range);
            
            const cached = this.getFromCache(cacheKey);
            if (cached) {
                console.log('Yahoo Finance cache hit:', symbol, interval);
                return cached.slice(-limit); // Apply limit to cached data
            }

            const url = this.buildUrl(symbol, interval, range);
            console.log('Yahoo Finance API request:', url);
            console.log('Yahoo Finance headers:', JSON.stringify(this.headers));
            
            const response = await fetch(url, { headers: this.headers });
            console.log('Yahoo Finance response status:', response.status, response.statusText);
            console.log('Yahoo Finance response headers:', JSON.stringify(Object.fromEntries(response.headers.entries())));
            
            if (!response.ok) {
                const errorText = await response.text();
                console.log('Yahoo Finance error response body:', errorText);
                throw new Error(`Yahoo Finance API error: ${response.status} ${response.statusText}`);
            }
            
            const responseText = await response.text();
            console.log('Yahoo Finance response body length:', responseText.length);
            console.log('Yahoo Finance response body start:', responseText.substring(0, 200));
            
            const data = JSON.parse(responseText);
            
            if (!data.chart || !data.chart.result || data.chart.result.length === 0) {
                console.warn('No chart data from Yahoo Finance for:', symbol);
                return [];
            }
            
            const result = data.chart.result[0];
            if (!result.timestamp || !result.indicators || !result.indicators.quote || !result.indicators.quote[0]) {
                console.warn('Invalid data structure from Yahoo Finance for:', symbol);
                return [];
            }
            
            const timestamps = result.timestamp;
            const quote = result.indicators.quote[0];
            const { open, high, low, close, volume } = quote;
            
            if (!timestamps || timestamps.length === 0) {
                console.warn('No timestamps in Yahoo Finance data for:', symbol);
                return [];
            }
            
            const convertedData = [];
            for (let i = 0; i < timestamps.length; i++) {
                if (open[i] !== null && high[i] !== null && low[i] !== null && close[i] !== null) {
                    convertedData.push(this.convertYahooCandle(
                        timestamps[i],
                        open[i],
                        high[i], 
                        low[i],
                        close[i],
                        volume[i]
                    ));
                }
            }
            
            // Sort by time ascending
            convertedData.sort((a, b) => a.openTime - b.openTime);
            
            this.setCache(cacheKey, convertedData);
            console.log(`Yahoo Finance data retrieved: ${convertedData.length} candles for ${symbol}`);
            
            // Apply limit
            return convertedData.slice(-limit);
            
        } catch (error) {
            console.error(`Yahoo Finance provider error for ${symbol}:`, error.message);
            if (error.stack) {
                console.error('Yahoo Finance error stack:', error.stack);
            }
            return [];
        }
    }
}