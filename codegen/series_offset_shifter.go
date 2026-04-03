package codegen

import (
	"regexp"
	"strconv"
	"strings"
)

var barFieldPrevAccess = map[string]string{
	"bar.Close":  "ctx.Data[i-1].Close",
	"bar.Open":   "ctx.Data[i-1].Open",
	"bar.High":   "ctx.Data[i-1].High",
	"bar.Low":    "ctx.Data[i-1].Low",
	"bar.Volume": "ctx.Data[i-1].Volume",
}

var seriesGetOffsetRe = regexp.MustCompile(`(\.Get\()(\d+)(\))`)

// historicalOHLCVAccessRe matches the bounds-checked IIFE emitted by GenerateHistoricalAccess
// for OHLCV fields, e.g.: func() float64 { if i-2 >= 0 { return ctx.Data[i-2].Close }; return math.NaN() }()
var historicalOHLCVAccessRe = regexp.MustCompile(
	`(func\(\) float64 \{ if i-)(\d+)( >= 0 \{ return ctx\.Data\[i-)(\d+)(\]\.\w+ \}; return math\.NaN\(\) \}\(\))`,
)

type SeriesOffsetShifter struct{}

func (SeriesOffsetShifter) ShiftToPrevBar(access string) string {
	if prev, ok := barFieldPrevAccess[access]; ok {
		return prev
	}
	if strings.Contains(access, "Series.GetCurrent()") {
		return strings.ReplaceAll(access, "Series.GetCurrent()", "Series.Get(1)")
	}
	if strings.Contains(access, ".Get(") {
		return seriesGetOffsetRe.ReplaceAllStringFunc(access, incrementSeriesGetOffset)
	}
	if strings.Contains(access, "ctx.Data[i-") {
		return historicalOHLCVAccessRe.ReplaceAllStringFunc(access, incrementHistoricalOHLCVOffset)
	}
	return access
}

func incrementSeriesGetOffset(match string) string {
	parts := seriesGetOffsetRe.FindStringSubmatch(match)
	n, _ := strconv.Atoi(parts[2])
	return parts[1] + strconv.Itoa(n+1) + parts[3]
}

func incrementHistoricalOHLCVOffset(match string) string {
	parts := historicalOHLCVAccessRe.FindStringSubmatch(match)
	n1, _ := strconv.Atoi(parts[2])
	n2, _ := strconv.Atoi(parts[4])
	return parts[1] + strconv.Itoa(n1+1) + parts[3] + strconv.Itoa(n2+1) + parts[5]
}
