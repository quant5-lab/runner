# Golden PineScript Strategies

---

## Strategy Catalog

| # | ✓ | Strategy | Complexity | Key Features | Source |
|---|---|----------|------------|--------------|--------|
| 1 | ☐ | **Supertrend Strategy** | Medium | ATR trailing stop, trend reversal signals | [GitHub](https://github.com/Alorse/pinescript-strategies) / [TradingView](https://www.tradingview.com/script/P5Cz5OPa-Supertrend/) |
| 2 | ☐ | **Bollinger Bands + RSI** | Medium | Band breakout + RSI filter, mean reversion | [GitHub](https://github.com/everget/tradingview-pinescript-indicators) |
| 3 | ☐ | **MACD Crossover Strategy** | Low | Signal line crossover, histogram divergence | [GitHub](https://github.com/Alorse/pinescript-strategies) |
| 4 | ☐ | **EMA Crossover (Triple EMA)** | Low | 3 EMA cross system, trend confirmation | [GitHub](https://github.com/800cherries/Tradingview-Indicators) |
| 5 | ☐ | **Ichimoku Cloud Strategy** | High | Cloud breakout, TK cross, Chikou confirmation | [GitHub](https://github.com/everget/tradingview-pinescript-indicators) |
| 6 | ☐ | **ADX + DI Strategy** | Medium | Trend strength filter, +DI/-DI crossover | [GitHub](https://github.com/Alorse/pinescript-strategies) |
| 7 | ☐ | **RSI Divergence Strategy** | Medium | Bull/bear divergence detection | [GitHub](https://github.com/just-nilux/awesome-tradingview) |
| 8 | ☐ | **Supply & Demand Zones** | High | Institutional zone detection, order blocks | [GitHub](https://github.com/800cherries/Tradingview-Indicators) |
| 9 | ☐ | **Volume Weighted Strategy** | Medium | VWAP deviation, volume confirmation | [GitHub](https://github.com/everget/tradingview-pinescript-indicators) |
| 10 | ☐ | **Keltner Channel Squeeze** | High | Volatility squeeze, momentum breakout | [GitHub](https://github.com/just-nilux/awesome-tradingview) |
| 11 | ☐ | **Pivot Points Reversal** | Medium | Support/resistance pivots, bounce plays | [GitHub](https://github.com/800cherries/Tradingview-Indicators) |
| 12 | ☐ | **Multi-Timeframe Confirmation** | High | HTF trend + LTF entry, MTF alignment | [GitHub](https://github.com/Alorse/pinescript-strategies) |

---

## Primary GitHub Repositories

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│ ✓ │ REPOSITORY                                    │ STARS │ STRATEGIES/INDICATORS│
├──────────────────────────────────────────────────────────────────────────────────┤
│ ☐ │ everget/tradingview-pinescript-indicators     │  754  │ 50+ indicators       │
│ ☐ │ Alorse/pinescript-strategies                  │  124  │ 50+ strategies       │
│ ☐ │ just-nilux/awesome-tradingview                │  369  │ Curated collection   │
│ ☐ │ 800cherries/Tradingview-Indicators            │  113  │ SMC/ICT strategies   │
│ ☐ │ pAulseperformance/awesome-pinescript          │  400+ │ Comprehensive list   │
│ ☐ │ Heavy91/TradingView_Indicators                │  265  │ Multiple strategies  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## Direct Source Links

### Complete Strategy Source Code

1. ☐ **Alorse Strategies Collection**
   - URL: https://github.com/Alorse/pinescript-strategies/tree/master/strategies
   - Contains: 50+ complete `.pine` strategy files

2. ☐ **800cherries SMC Strategies**
   - URL: https://github.com/800cherries/Tradingview-Indicators/tree/main/strategies
   - Contains: Institutional-style strategies (Supply/Demand, Pivots)

3. ☐ **Everget Indicators**
   - URL: https://github.com/everget/tradingview-pinescript-indicators
   - Contains: Oscillators, bands, moving averages, volatility indicators

4. ☐ **PineCoders Utils**
   - URL: https://github.com/pinecoders/pine-utils
   - Contains: Templates, boilerplates, utility functions

5. ☐ **Awesome TradingView**
   - URL: https://github.com/just-nilux/awesome-tradingview
   - Contains: Curated list with direct links to strategies

---

## TradingView Direct Links (Open Source Scripts)

| ✓ | Strategy Type | TradingView Link |
|---|--------------|------------------|
| ☐ | Supertrend | https://www.tradingview.com/scripts/supertrend/ |
| ☐ | Bollinger Bands | https://www.tradingview.com/scripts/bollingerbands/ |
| ☐ | MACD | https://www.tradingview.com/scripts/macd/ |
| ☐ | RSI | https://www.tradingview.com/scripts/relativestrengthindex/ |
| ☐ | Ichimoku | https://www.tradingview.com/scripts/ichimoku/ |
| ☐ | Moving Average | https://www.tradingview.com/scripts/movingaverage/ |
| ☐ | Volume Profile | https://www.tradingview.com/scripts/volumeprofile/ |
| ☐ | ADX | https://www.tradingview.com/scripts/adx/ |

---

## Pine Script Features Coverage Matrix

```
Feature                    │ Required │ Coverage
───────────────────────────┼──────────┼─────────
ta.sma/ema/wma             │    ✓     │ All strategies
ta.rsi                     │    ✓     │ #2, #7
ta.macd                    │    ✓     │ #3
ta.atr                     │    ✓     │ #1, #10
ta.adx                     │    ✓     │ #6
ta.bb                      │    ✓     │ #2, #10
ta.supertrend              │    ✓     │ #1
ta.pivothigh/pivotlow      │    ✓     │ #8, #11
ta.crossover/crossunder    │    ✓     │ All strategies
request.security (MTF)     │    ✓     │ #12
strategy.entry/exit        │    ✓     │ All strategies
strategy.close             │    ✓     │ All strategies
alertcondition             │    ✓     │ #1-#12
array operations           │    ✓     │ #8
line/box/label             │    ✓     │ #8, #11
```

---

## Validation Criteria

For runner to pass golden milestone:

1. **Indicator Calculation** - All `ta.*` functions produce matching values
2. **Signal Generation** - Entry/exit signals match TradingView backtest
3. **Historical Access** - `[n]` lookback works correctly (ForwardSeriesBuffer)
4. **Multi-Timeframe** - `request.security` returns correct HTF data
5. **Strategy Execution** - Position management matches expected behavior

---

## Implementation Priority

```
Phase 1: Basic Strategies (Low complexity)
☐ MACD Crossover (#3)
☐ EMA Crossover (#4)
☐ RSI Strategy (basic)

Phase 2: Intermediate Strategies
☐ Supertrend (#1)
☐ Bollinger Bands + RSI (#2)
☐ ADX + DI (#6)
☐ Volume Weighted (#9)

Phase 3: Advanced Strategies
☐ Ichimoku Cloud (#5)
☐ Supply & Demand Zones (#8)
☐ Keltner Squeeze (#10)
☐ MTF Confirmation (#12)
```

---

## References

- TradingView Scripts: https://www.tradingview.com/scripts/
- Pine Script v5 Docs: https://www.tradingview.com/pine-script-docs/
- GitHub Topic: https://github.com/topics/pinescript
- PineCoders: https://www.pinecoders.com/

---

## PROMPT

``````
Enhance golden reference testing mechanism by putting the .pine code of strategy into the codebase, and then updating it's reference data, establishing a regression test for it
Strategy name: `Bollinger Bands + RSI` from `golden-strategies.md`

In case when strategy fails to run, report this immediately and stop further operation until user explicit proceeding approval
``````