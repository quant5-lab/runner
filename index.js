import { PineTS, Provider } from '../PineTS/dist/pinets.dev.es.js';

console.log('🚀 Starting PineTS Example App...');

async function runEMAExample() {
    try {
        console.log('📊 Initializing PineTS with market data...');
        
        // Initialize with market data (Binance BTCUSDT daily data, 100 bars)
        const pineTS = new PineTS(Provider.Binance, 'BTCUSDT', 'D', 100);
        
        console.log('⚡ Running EMA crossover indicator...');
        
        // Run EMA crossover indicator based on the documentation example
        const { result } = await pineTS.run((context) => {
            const ta = context.ta;
            const math = context.math;
            const { close, open } = context.data;
            
            console.log('📈 Calculating EMAs...');
            
            // Calculate EMAs
            const ema9 = ta.ema(close, 9);
            const ema18 = ta.ema(close, 18);
            
            // Determine bias
            const bull_bias = ema9 > ema18;
            const bear_bias = ema9 < ema18;
            
            // Get previous close and difference
            const prev_close = close[1];
            const diff_close = close - prev_close;
            
            // Some additional calculations
            const abs_diff = math.abs(open[1] - close[2]);
            
            console.log('✅ Calculations complete!');
            
            // Return the results
            return {
                ema9,
                ema18,
                bull_bias,
                bear_bias,
                prev_close,
                diff_close,
                abs_diff
            };
        });
        
        console.log('📋 Results:');
        console.log('- EMA 9 (last 3 values):', result.ema9.slice(-3));
        console.log('- EMA 18 (last 3 values):', result.ema18.slice(-3));
        console.log('- Bull Bias (last 3 values):', result.bull_bias.slice(-3));
        console.log('- Bear Bias (last 3 values):', result.bear_bias.slice(-3));
        console.log('- Previous Close (last 3 values):', result.prev_close.slice(-3));
        console.log('- Close Difference (last 3 values):', result.diff_close.slice(-3));
        console.log('- Absolute Difference (last 3 values):', result.abs_diff.slice(-3));
        
        console.log('🎉 PineTS Example completed successfully!');
        
    } catch (error) {
        console.error('❌ Error running PineTS example:', error);
        process.exit(1);
    }
}

// Run the example
runEMAExample();