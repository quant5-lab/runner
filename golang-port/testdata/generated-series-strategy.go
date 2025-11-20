package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/borisquantlab/pinescript-go/runtime/chartdata"
	"github.com/borisquantlab/pinescript-go/runtime/context"
	"github.com/borisquantlab/pinescript-go/runtime/output"
	"github.com/borisquantlab/pinescript-go/runtime/series"
	"github.com/borisquantlab/pinescript-go/runtime/strategy"
	"github.com/borisquantlab/pinescript-go/runtime/ta"
	_ "github.com/borisquantlab/pinescript-go/runtime/value" // May be used by generated code
)

/* CLI flags */
var (
	symbolFlag    = flag.String("symbol", "", "Trading symbol (e.g., BTCUSDT)")
	timeframeFlag = flag.String("timeframe", "1h", "Timeframe (e.g., 1m, 5m, 1h, 1D)")
	dataFlag      = flag.String("data", "", "Path to OHLCV data JSON file")
	outputFlag    = flag.String("output", "chart-data.json", "Output file path")
)

/* Strategy execution function - INJECTED BY CODEGEN */
func executeStrategy(ctx *context.Context) (*output.Collector, *strategy.Strategy) {
	collector := output.NewCollector()
	strat := strategy.NewStrategy()

	strat.Call("Generated Strategy", 10000)

	// ALL variables use Series storage (ForwardSeriesBuffer paradigm)
	var prev_sma20Series *series.Series
	var crossover_signalSeries *series.Series
	var ta_crossoverSeries *series.Series
	var manual_signalSeries *series.Series
	var ta_signalSeries *series.Series
	var sma20Series *series.Series
	var sma50Series *series.Series
	var prev_sma50Series *series.Series
	var crossunder_signalSeries *series.Series
	var ta_crossunderSeries *series.Series

	// Initialize Series storage
	prev_sma50Series = series.NewSeries(len(ctx.Data))
	crossunder_signalSeries = series.NewSeries(len(ctx.Data))
	ta_crossunderSeries = series.NewSeries(len(ctx.Data))
	prev_sma20Series = series.NewSeries(len(ctx.Data))
	crossover_signalSeries = series.NewSeries(len(ctx.Data))
	ta_crossoverSeries = series.NewSeries(len(ctx.Data))
	manual_signalSeries = series.NewSeries(len(ctx.Data))
	ta_signalSeries = series.NewSeries(len(ctx.Data))
	sma20Series = series.NewSeries(len(ctx.Data))
	sma50Series = series.NewSeries(len(ctx.Data))

	// Pre-calculate TA functions using runtime library
	closeSeries := make([]float64, len(ctx.Data))
	for i := range ctx.Data {
		closeSeries[i] = ctx.Data[i].Close
	}

	sma20Array := ta.Sma(closeSeries, 20)
	sma50Array := ta.Sma(closeSeries, 50)

	for i := 0; i < len(ctx.Data); i++ {
		ctx.BarIndex = i
		bar := ctx.Data[i]
		strat.OnBarUpdate(i, bar.Open, bar.Time)

		sma20Series.Set(sma20Array[i])
		sma50Series.Set(sma50Array[i])
		prev_sma20Series.Set(sma20Series.Get(1))
		prev_sma50Series.Set(sma50Series.Get(1))
		crossover_signalSeries.Set(func() float64 {
			if sma20Series.Get(0) > sma50Series.Get(0) && prev_sma20Series.Get(0) <= prev_sma50Series.Get(0) {
				return 1.0
			} else {
				return 0.0
			}
		}())
		crossunder_signalSeries.Set(func() float64 {
			if sma20Series.Get(0) < sma50Series.Get(0) && prev_sma20Series.Get(0) >= prev_sma50Series.Get(0) {
				return 1.0
			} else {
				return 0.0
			}
		}())
		// Crossover: sma20Series.Get(0) crosses above sma50Series.Get(0)
		if i > 0 {
			ta_crossover_prev1 := sma20Series.Get(1)
			ta_crossover_prev2 := sma50Series.Get(1)
			ta_crossoverSeries.Set(func() float64 {
				if sma20Series.Get(0) > sma50Series.Get(0) && ta_crossover_prev1 <= ta_crossover_prev2 {
					return 1.0
				} else {
					return 0.0
				}
			}())
		} else {
			ta_crossoverSeries.Set(0.0)
		}
		// Crossunder: sma20Series.Get(0) crosses below sma50Series.Get(0)
		if i > 0 {
			ta_crossunder_prev1 := sma20Series.Get(1)
			ta_crossunder_prev2 := sma50Series.Get(1)
			ta_crossunderSeries.Set(func() float64 {
				if sma20Series.Get(0) < sma50Series.Get(0) && ta_crossunder_prev1 >= ta_crossunder_prev2 {
					return 1.0
				} else {
					return 0.0
				}
			}())
		} else {
			ta_crossunderSeries.Set(0.0)
		}
		manual_signalSeries.Set(func() float64 {
			if crossover_signalSeries.Get(0) != 0 {
				return 1.00
			} else {
				return 0.00
			}
		}())
		ta_signalSeries.Set(func() float64 {
			if ta_crossoverSeries.Get(0) != 0 {
				return 1.00
			} else {
				return 0.00
			}
		}())
		if crossover_signalSeries.Get(0) != 0 {
			strat.Entry("Long", strategy.Long, 1)
		}
		if crossunder_signalSeries.Get(0) != 0 {
			strat.Entry("Short", strategy.Short, 1)
		}
		collector.Add("sma20", bar.Time, sma20Series.Get(0), nil)
		collector.Add("sma50", bar.Time, sma50Series.Get(0), nil)
		collector.Add("manual_signal", bar.Time, manual_signalSeries.Get(0), nil)
		collector.Add("ta_signal", bar.Time, ta_signalSeries.Get(0), nil)

		// Suppress unused variable warnings
		_ = ta_signalSeries
		_ = sma20Series
		_ = sma50Series
		_ = prev_sma50Series
		_ = crossunder_signalSeries
		_ = ta_crossunderSeries
		_ = prev_sma20Series
		_ = crossover_signalSeries
		_ = ta_crossoverSeries
		_ = manual_signalSeries

		// Advance Series cursors
		if i < len(ctx.Data)-1 {
			sma20Series.Next()
		}
		if i < len(ctx.Data)-1 {
			sma50Series.Next()
		}
		if i < len(ctx.Data)-1 {
			prev_sma50Series.Next()
		}
		if i < len(ctx.Data)-1 {
			crossunder_signalSeries.Next()
		}
		if i < len(ctx.Data)-1 {
			ta_crossunderSeries.Next()
		}
		if i < len(ctx.Data)-1 {
			prev_sma20Series.Next()
		}
		if i < len(ctx.Data)-1 {
			crossover_signalSeries.Next()
		}
		if i < len(ctx.Data)-1 {
			ta_crossoverSeries.Next()
		}
		if i < len(ctx.Data)-1 {
			manual_signalSeries.Next()
		}
		if i < len(ctx.Data)-1 {
			ta_signalSeries.Next()
		}
	}

	return collector, strat
}

func main() {
	flag.Parse()

	if *symbolFlag == "" || *dataFlag == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -symbol SYMBOL -data DATA.json [-timeframe 1h] [-output chart-data.json]\n", os.Args[0])
		os.Exit(1)
	}

	/* Load OHLCV data */
	dataBytes, err := os.ReadFile(*dataFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read data file: %v\n", err)
		os.Exit(1)
	}

	var bars []context.OHLCV
	err = json.Unmarshal(dataBytes, &bars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse data JSON: %v\n", err)
		os.Exit(1)
	}

	if len(bars) == 0 {
		fmt.Fprintf(os.Stderr, "No bars in data file\n")
		os.Exit(1)
	}

	/* Create runtime context */
	ctx := context.New(*symbolFlag, *timeframeFlag, len(bars))
	for _, bar := range bars {
		ctx.AddBar(bar)
	}

	/* Execute strategy */
	startTime := time.Now()
	plotCollector, strat := executeStrategy(ctx)
	executionTime := time.Since(startTime)

	/* Generate chart data with metadata */
	cd := chartdata.NewChartData(ctx, *symbolFlag, *timeframeFlag, "Generated Strategy")
	cd.AddPlots(plotCollector)
	cd.AddStrategy(strat, ctx.Data[len(ctx.Data)-1].Close)

	/* Write output */
	jsonBytes, err := cd.ToJSON()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(*outputFlag, jsonBytes, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
		os.Exit(1)
	}

	/* Print summary */
	fmt.Printf("Symbol: %s\n", *symbolFlag)
	fmt.Printf("Timeframe: %s\n", *timeframeFlag)
	fmt.Printf("Bars: %d\n", len(bars))
	fmt.Printf("Execution time: %v\n", executionTime)
	fmt.Printf("Output: %s (%d bytes)\n", *outputFlag, len(jsonBytes))

	if strat != nil {
		th := strat.GetTradeHistory()
		closedTrades := th.GetClosedTrades()
		fmt.Printf("Closed trades: %d\n", len(closedTrades))
		fmt.Printf("Final equity: %.2f\n", strat.GetEquity(ctx.Data[len(ctx.Data)-1].Close))
	}
}
