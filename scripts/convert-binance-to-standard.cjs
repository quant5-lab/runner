#!/usr/bin/env node
// Convert Binance format to standard OHLCV format
// Usage: node scripts/convert-binance-to-standard.js input.json output.json [metadata.json]

const fs = require('fs');

const [,, inputFile, outputFile, metadataFile] = process.argv;

if (!inputFile || !outputFile) {
    console.error('Usage: node convert-binance-to-standard.js input.json output.json [metadata.json]');
    process.exit(1);
}

const binanceData = JSON.parse(fs.readFileSync(inputFile, 'utf8'));

// Check if binanceData has a 'data' field (provider result object) or is an array (raw data)
const barsArray = Array.isArray(binanceData) ? binanceData : binanceData.data || binanceData;

const standardData = barsArray.map(bar => ({
    time: Math.floor(bar.openTime / 1000), // Convert ms to seconds
    open: parseFloat(bar.open),
    high: parseFloat(bar.high),
    low: parseFloat(bar.low),
    close: parseFloat(bar.close),
    volume: parseFloat(bar.volume)
}));

// If metadata file is provided, add timezone to the output
if (metadataFile && fs.existsSync(metadataFile)) {
    const metadata = JSON.parse(fs.readFileSync(metadataFile, 'utf8'));
    const outputWithMetadata = {
        timezone: metadata.timezone || 'UTC',
        bars: standardData
    };
    fs.writeFileSync(outputFile, JSON.stringify(outputWithMetadata, null, 2));
    console.log(`Converted ${standardData.length} bars with timezone ${metadata.timezone}: ${inputFile} → ${outputFile}`);
} else {
    fs.writeFileSync(outputFile, JSON.stringify(standardData, null, 2));
    console.log(`Converted ${standardData.length} bars: ${inputFile} → ${outputFile}`);
}

