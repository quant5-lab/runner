#!/usr/bin/env node
// Convert Binance format to standard OHLCV format
// Usage: node scripts/convert-binance-to-standard.js input.json output.json

const fs = require('fs');

const [,, inputFile, outputFile] = process.argv;

if (!inputFile || !outputFile) {
    console.error('Usage: node convert-binance-to-standard.js input.json output.json');
    process.exit(1);
}

const binanceData = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

const standardData = binanceData.map(bar => ({
    time: Math.floor(bar.openTime / 1000), // Convert ms to seconds
    open: parseFloat(bar.open),
    high: parseFloat(bar.high),
    low: parseFloat(bar.low),
    close: parseFloat(bar.close),
    volume: parseFloat(bar.volume)
}));

fs.writeFileSync(outputFile, JSON.stringify(standardData, null, 2));
console.log(`Converted ${standardData.length} bars: ${inputFile} → ${outputFile}`);
