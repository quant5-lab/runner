package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Bar struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

type MarketData struct {
	Symbol    string `json:"symbol"`
	Timeframe string `json:"timeframe"`
	Period    string `json:"period"`
	Timezone  string `json:"timezone"`
	Bars      []Bar  `json:"bars"`
}

func main() {
	symbol := flag.String("symbol", "SPY", "Symbol name")
	timeframe := flag.String("timeframe", "1h", "Timeframe (1h, D, W, M)")
	bars := flag.Int("bars", 500, "Number of bars")
	output := flag.String("output", "", "Output file path")
	flag.Parse()

	if *output == "" {
		log.Fatal("Output file path required (-output)")
	}

	data := generateSyntheticData(*symbol, *timeframe, *bars)

	if err := saveJSON(*output, data); err != nil {
		log.Fatalf("Save JSON: %v", err)
	}

	fmt.Printf("Generated %d bars for %s %s -> %s\n", *bars, *symbol, *timeframe, *output)
}

func generateSyntheticData(symbol, timeframe string, barCount int) *MarketData {
	bars := make([]Bar, barCount)

	basePrice := 100.0
	baseTime := int64(1609772400)
	timeInterval := getTimeInterval(timeframe)

	currentPrice := basePrice
	trendStrength := 0.0
	volatilityRegime := 1.0

	for i := 0; i < barCount; i++ {
		if i%200 == 0 {
			trendStrength = float64((i/200)%3-1) * 0.15
			volatilityRegime = []float64{0.5, 1.0, 2.0, 1.5}[(i/200)%4]
		}

		drift := trendStrength * pseudoRandom(i, 0)
		noise := pseudoRandom(i, 1) * 0.8 * volatilityRegime
		priceChange := drift + noise

		if i%137 == 0 {
			priceChange *= 2.5
		}

		open := currentPrice
		close := currentPrice * (1 + priceChange/100.0)

		barVolatility := volatilityRegime * (0.3 + pseudoRandom(i, 2)*0.7)
		high := max(open, close) * (1 + barVolatility/100.0)
		low := min(open, close) * (1 - barVolatility/100.0)

		bars[i] = Bar{
			Time:   baseTime + int64(i)*timeInterval,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: 1000000 * (1 + volatilityRegime*0.5) * (0.8 + pseudoRandom(i, 3)*0.4),
		}

		currentPrice = close
	}

	return &MarketData{
		Symbol:    symbol,
		Timeframe: timeframe,
		Period:    fmt.Sprintf("synthetic-%d-bars", barCount),
		Timezone:  "America/New_York",
		Bars:      bars,
	}
}

func getTimeInterval(timeframe string) int64 {
	switch timeframe {
	case "1h":
		return 3600
	case "D":
		return 86400
	case "W":
		return 604800
	case "M":
		return 2592000
	default:
		return 3600
	}
}

// pseudoRandom generates deterministic pseudo-random values using simple hashing
func pseudoRandom(seed int, salt int) float64 {
	x := (seed*2654435761 + salt*1103515245) & 0x7FFFFFFF
	return (float64(x)/float64(0x7FFFFFFF))*2 - 1
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func saveJSON(path string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
