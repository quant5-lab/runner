package tv_reference

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TVTrade struct {
	Number     int
	Direction  string
	EntryUTC   time.Time
	EntryPrice float64
	ExitUTC    time.Time
	ExitPrice  float64
	Size       float64
	NetPnL     float64
}

type TVTimezone int

const (
	TVTimezoneUTC TVTimezone = iota
	TVTimezoneMoscow
	TVTimezoneNewYork
)

type RunnerTrade struct {
	EntryUTC   time.Time
	EntryPrice float64
	Direction  string
	Size       float64
	NetPnL     float64
}

const SizeResidualDenominatorFloor = 1.0

func LoadTrades(path string, tz TVTimezone) ([]TVTrade, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open TV CSV %s: %w", path, err)
	}
	defer f.Close()
	return parseTrades(f, tz)
}

func FilterByEntryWindow(trades []TVTrade, start, end time.Time) []TVTrade {
	var out []TVTrade
	for _, t := range trades {
		if !t.EntryUTC.Before(start) && !t.EntryUTC.After(end) {
			out = append(out, t)
		}
	}
	return out
}

func MatchExact(runner []RunnerTrade, tv []TVTrade, timeTol time.Duration, priceTol float64) (matched, runnerOnly, tvOnly int) {
	orderedRunner := append([]RunnerTrade(nil), runner...)
	orderedTV := append([]TVTrade(nil), tv...)
	sort.Slice(orderedRunner, func(i, j int) bool { return orderedRunner[i].EntryUTC.Before(orderedRunner[j].EntryUTC) })
	sort.Slice(orderedTV, func(i, j int) bool { return orderedTV[i].EntryUTC.Before(orderedTV[j].EntryUTC) })

	usedTV := make([]bool, len(orderedTV))
	for _, rt := range orderedRunner {
		k := firstMatchingTVTrade(rt, orderedTV, usedTV, timeTol, priceTol)
		if k < 0 {
			runnerOnly++
			continue
		}
		usedTV[k] = true
		matched++
	}
	tvOnly = len(orderedTV) - matched
	return
}

// TVEquityAtWindowStart returns the strategy equity at the start of windowStart,
// computed as initialCapital plus the cumulative net PnL of every trade that
// entered before windowStart. Pass the full (unfiltered) trade slice so that
// pre-window trades are included in the sum.
func TVEquityAtWindowStart(allTrades []TVTrade, windowStart time.Time, initialCapital float64) float64 {
	equity := initialCapital
	for _, t := range allTrades {
		if t.EntryUTC.Before(windowStart) {
			equity += t.NetPnL
		}
	}
	return equity
}

// MatchNetPnL pairs each runner closed trade to a TV trade by entry time,
// direction, and entry price (same tolerances as MatchExact), then checks
// each matched pair's net PnL within relTol (relative).
//
// equityRatio scales runner PnL to the TV equity basis before comparison:
//
//	equityRatio = tvEquityAtWindowStart / runnerInitialCapital
//
// Pass 1.0 when runner and TV share the same starting equity. When the TV
// history predates the fixture window the runner starts at a higher nominal
// equity; multiplying runner PnL by equityRatio normalises it to the TV equity
// basis, making the comparison equity-invariant for percent-of-equity strategies.
//
// TV trades that have no matching runner trade (e.g. still-open positions) are
// silently skipped; runner trades with no TV match increment pnlMismatch.
// Returns the count of matched pairs and the count whose scaled PnL deviates
// beyond relTol.
func MatchNetPnL(runner []RunnerTrade, tv []TVTrade, equityRatio float64, timeTol time.Duration, priceTol, relTol float64) (matched, pnlMismatch int) {
	usedTV := make([]bool, len(tv))
	for _, rt := range runner {
		k := firstMatchingTVTrade(rt, tv, usedTV, timeTol, priceTol)
		if k < 0 {
			continue
		}
		usedTV[k] = true
		matched++
		scaled := rt.NetPnL * equityRatio
		if !pnlMatchRelative(scaled, tv[k].NetPnL, relTol) {
			pnlMismatch++
		}
	}
	return
}

func MatchSize(runner []RunnerTrade, tv []TVTrade, timeTol time.Duration, priceTol, relTol float64) (matched, sizeMismatch int, maxResidual float64) {
	usedTV := make([]bool, len(tv))
	for _, rt := range runner {
		k := firstMatchingTVTrade(rt, tv, usedTV, timeTol, priceTol)
		if k < 0 {
			continue
		}
		usedTV[k] = true
		matched++
		residual := SizeResidual(rt.Size, tv[k].Size)
		if residual > maxResidual {
			maxResidual = residual
		}
		if residual > relTol {
			sizeMismatch++
		}
	}
	return
}

func SizeResidual(runnerSize, tvSize float64) float64 {
	if runnerSize == tvSize {
		return 0
	}
	denominator := mathMax(floatAbs(tvSize), SizeResidualDenominatorFloor)
	return floatAbs(runnerSize-tvSize) / denominator
}

// HasSizeData reports whether any trade in the slice carries a nonzero Size,
// indicating that the source CSV contained a size column.
func HasSizeData(trades []TVTrade) bool {
	for _, t := range trades {
		if t.Size != 0 {
			return true
		}
	}
	return false
}

func pnlMatchRelative(a, b, relTol float64) bool {
	if a == b {
		return true
	}
	scale := a
	if floatAbs(b) > floatAbs(a) {
		scale = b
	}
	if scale == 0 {
		return true
	}
	return floatAbs(a-b)/floatAbs(scale) <= relTol
}

func firstMatchingTVTrade(rt RunnerTrade, tv []TVTrade, used []bool, timeTol time.Duration, priceTol float64) int {
	for i, tt := range tv {
		if used[i] {
			continue
		}
		if tradesMatch(rt, tt, timeTol, priceTol) {
			return i
		}
	}
	return -1
}

func tradesMatch(rt RunnerTrade, tt TVTrade, timeTol time.Duration, priceTol float64) bool {
	return durationAbs(tt.EntryUTC.Sub(rt.EntryUTC)) <= timeTol &&
		floatAbs(tt.EntryPrice-rt.EntryPrice) <= priceTol &&
		directionsMatch(rt.Direction, tt.Direction)
}

func durationAbs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func floatAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func directionsMatch(runnerDirection, tvDirection string) bool {
	runnerDirection, runnerOK := normalizeDirection(runnerDirection)
	tvDirection, tvOK := normalizeDirection(tvDirection)
	return runnerOK && tvOK && runnerDirection == tvDirection
}

func normalizeDirection(direction string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(direction)) {
	case "long":
		return "long", true
	case "short":
		return "short", true
	default:
		return "", false
	}
}

type csvRow struct {
	num     int
	rowType string
	dt      time.Time
	price   float64
	size    float64
	netPnL  float64
}

type halfTrade struct {
	dt     time.Time
	price  float64
	size   float64
	typ    string
	netPnL float64
}

type tradePair struct {
	entry *halfTrade
	exit  *halfTrade
}

func parseTrades(r io.Reader, tz TVTimezone) ([]TVTrade, error) {
	cr := csv.NewReader(r)
	cr.LazyQuotes = true

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read CSV header: %w", err)
	}

	col := indexColumns(header)
	if col.tradeNum < 0 || col.rowType < 0 || col.datetime < 0 || col.price < 0 {
		return nil, fmt.Errorf("CSV missing required columns; got: %v", header)
	}

	var rows []csvRow
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read CSV row: %w", err)
		}
		if len(rec) <= col.price {
			continue
		}
		num, err := strconv.Atoi(strings.TrimSpace(rec[col.tradeNum]))
		if err != nil {
			continue
		}
		dt, err := parseDateTime(strings.TrimSpace(rec[col.datetime]), tz)
		if err != nil {
			continue
		}
		price, err := strconv.ParseFloat(strings.TrimSpace(rec[col.price]), 64)
		if err != nil {
			continue
		}
		var netPnL float64
		if col.netPnL >= 0 && col.netPnL < len(rec) {
			netPnL, _ = strconv.ParseFloat(strings.TrimSpace(rec[col.netPnL]), 64)
		}
		var size float64
		if col.size >= 0 && col.size < len(rec) {
			size, _ = strconv.ParseFloat(strings.TrimSpace(rec[col.size]), 64)
		}
		rows = append(rows, csvRow{num, strings.TrimSpace(rec[col.rowType]), dt, price, size, netPnL})
	}

	return assembleTrades(rows), nil
}

func assembleTrades(rows []csvRow) []TVTrade {
	byNum := map[int]*tradePair{}
	var order []int
	for _, row := range rows {
		if _, ok := byNum[row.num]; !ok {
			byNum[row.num] = &tradePair{}
			order = append(order, row.num)
		}
		p := byNum[row.num]
		h := &halfTrade{row.dt, row.price, row.size, row.rowType, row.netPnL}
		if isEntry(row.rowType) {
			p.entry = h
		} else if isExit(row.rowType) {
			p.exit = h
		}
	}

	var trades []TVTrade
	for _, num := range order {
		p := byNum[num]
		if p.entry == nil {
			continue
		}
		tr := TVTrade{
			Number:     num,
			Direction:  directionFromType(p.entry.typ),
			EntryUTC:   p.entry.dt,
			EntryPrice: p.entry.price,
			Size:       p.entry.size,
			NetPnL:     p.entry.netPnL,
		}
		if p.exit != nil {
			tr.ExitUTC = p.exit.dt
			tr.ExitPrice = p.exit.price
		}
		trades = append(trades, tr)
	}
	return trades
}

type colIndex struct {
	tradeNum int
	rowType  int
	datetime int
	price    int
	size     int
	netPnL   int
}

func indexColumns(header []string) colIndex {
	idx := colIndex{-1, -1, -1, -1, -1, -1}
	for i, h := range header {
		h = normalizeHeader(h)
		switch {
		case h == "Trade number":
			idx.tradeNum = i
		case h == "Type":
			idx.rowType = i
		case h == "Date and time":
			idx.datetime = i
		case strings.HasPrefix(h, "Price") && idx.price < 0:
			idx.price = i
		case isSizeHeader(h) && idx.size < 0:
			idx.size = i
		case strings.HasPrefix(h, "Net PnL") && idx.netPnL < 0:
			idx.netPnL = i
		}
	}
	return idx
}

func normalizeHeader(header string) string {
	return strings.TrimSpace(strings.TrimPrefix(header, "\xef\xbb\xbf"))
}

func isSizeHeader(header string) bool {
	normalized := strings.ToLower(strings.TrimSpace(header))
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	switch normalized {
	case "size", "quantity", "contracts", "size (qty)", "qty":
		return true
	default:
		return false
	}
}

var newYorkLocation = func() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.FixedZone("EST", -5*3600)
	}
	return loc
}()

func parseDateTime(s string, tz TVTimezone) (time.Time, error) {
	switch tz {
	case TVTimezoneMoscow:
		t, err := time.Parse("2006-01-02 15:04", s)
		if err != nil {
			return time.Time{}, err
		}
		return t.Add(-3 * time.Hour), nil
	case TVTimezoneNewYork:
		t, err := time.ParseInLocation("2006-01-02 15:04", s, newYorkLocation)
		if err != nil {
			return time.Time{}, err
		}
		return t.UTC(), nil
	case TVTimezoneUTC:
		t, err := time.Parse("2006-01-02 15:04", s)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	default:
		t, err := time.Parse("2006-01-02 15:04", s)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}
}

func isEntry(s string) bool { return strings.Contains(s, "Entry") }
func isExit(s string) bool  { return strings.Contains(s, "Exit") || strings.Contains(s, "Close") }

func directionFromType(s string) string {
	if strings.Contains(strings.ToLower(s), "long") {
		return "long"
	}
	return "short"
}
