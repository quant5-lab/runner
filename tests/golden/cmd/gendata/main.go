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
	baseTime := int64(1609459200)
	timeInterval := getTimeInterval(timeframe)

	for i := 0; i < barCount; i++ {
		trend := float64(i) * 0.05
		volatility := 2.0

		open := basePrice + trend + randomWalk(volatility)
		high := open + randomPositive(volatility)
		low := open - randomPositive(volatility)
		close := open + randomWalk(volatility)

		if high < open {
			high = open
		}
		if high < close {
			high = close
		}
		if low > open {
			low = open
		}
		if low > close {
			low = close
		}

		bars[i] = Bar{
			Time:   baseTime + int64(i)*timeInterval,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: 1000000 + float64(i%100)*10000,
		}
	}

	return &MarketData{
		Symbol:    symbol,
		Timeframe: timeframe,
		Period:    fmt.Sprintf("synthetic-%d-bars", barCount),
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

func randomWalk(magnitude float64) float64 {
	return (float64(len(os.Args)%10) - 5) * magnitude / 10
}

func randomPositive(magnitude float64) float64 {
	return float64(len(os.Args)%5+1) * magnitude / 5
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
