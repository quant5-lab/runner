#!/usr/bin/env python3
"""Generate synthetic OHLCV test data for security() testing"""

import json
import sys
from datetime import datetime, timedelta

def generate_synthetic_bars(start_timestamp, count, interval_seconds, base_price=100.0):
    """Generate synthetic OHLCV bars with trending price"""
    bars = []
    price = base_price
    
    for i in range(count):
        timestamp = start_timestamp + i * interval_seconds
        
        # Create trend: gradual increase with some volatility
        trend = i * 0.5  # Slow uptrend
        volatility = (i % 10 - 5) * 0.3  # Small oscillations
        price = base_price + trend + volatility
        
        # Generate OHLC with realistic spread
        open_price = price
        high = price + (abs(i % 7) * 0.5)
        low = price - (abs((i+3) % 5) * 0.4)
        close_price = price + ((i % 3 - 1) * 0.2)
        volume = 1000 + (i % 100) * 10
        
        bars.append({
            "time": timestamp,
            "open": round(open_price, 2),
            "high": round(high, 2),
            "low": round(low, 2),
            "close": round(close_price, 2),
            "volume": volume
        })
    
    return bars

def main():
    # Start timestamp: January 1, 2024
    start_date = datetime(2024, 1, 1, 0, 0, 0)
    start_timestamp = int(start_date.timestamp())
    
    # Generate 1h data (500 bars = ~20 days)
    print("Generating BTCUSDT_1h.json (500 bars)...", file=sys.stderr)
    hourly_bars = generate_synthetic_bars(
        start_timestamp=start_timestamp,
        count=500,
        interval_seconds=3600,  # 1 hour
        base_price=40000.0
    )
    
    with open('testdata/ohlcv/BTCUSDT_1h.json', 'w') as f:
        json.dump(hourly_bars, f, indent=2)
    print(f"Created testdata/ohlcv/BTCUSDT_1h.json ({len(hourly_bars)} bars)", file=sys.stderr)
    
    # Generate 1D data (250 bars = ~250 days for SMA200)
    print("Generating BTCUSDT_1D.json (250 bars)...", file=sys.stderr)
    daily_bars = generate_synthetic_bars(
        start_timestamp=start_timestamp,
        count=250,
        interval_seconds=86400,  # 1 day
        base_price=40000.0
    )
    
    with open('testdata/ohlcv/BTCUSDT_1D.json', 'w') as f:
        json.dump(daily_bars, f, indent=2)
    print(f"Created testdata/ohlcv/BTCUSDT_1D.json ({len(daily_bars)} bars)", file=sys.stderr)
    
    print("\nTest data generation complete!", file=sys.stderr)
    print(f"1h bars: {len(hourly_bars)}, 1D bars: {len(daily_bars)}", file=sys.stderr)

if __name__ == '__main__':
    main()
