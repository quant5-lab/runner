package codegen

import "fmt"

/* TR = max(high-low, |high-prevClose|, |low-prevClose|); at bar 0: high - low */
type TrueRangeAccessGenerator struct{}

func NewTrueRangeAccessGenerator() *TrueRangeAccessGenerator {
	return &TrueRangeAccessGenerator{}
}

func (g *TrueRangeAccessGenerator) GenerateLoopValueAccess(loopVar string) string {
	return fmt.Sprintf("func() float64 { idx := ctx.BarIndex - %s; h := ctx.Data[idx].High; l := ctx.Data[idx].Low; if idx == 0 { return h - l }; pc := ctx.Data[idx-1].Close; return math.Max(h-l, math.Max(math.Abs(h-pc), math.Abs(l-pc))) }()", loopVar)
}

func (g *TrueRangeAccessGenerator) GenerateInitialValueAccess(period int) string {
	return g.GenerateLoopValueAccess(fmt.Sprintf("%d", period-1))
}

func (g *TrueRangeAccessGenerator) GenerateCurrentValueAccess() string {
	return g.GenerateLoopValueAccess("0")
}

func (g *TrueRangeAccessGenerator) GetBaseOffset() int {
	return 0
}
