package codegen

import "fmt"

func emitWarmupGuard(g *generator, varName, warmupExpr string, bodyFn func() string) string {
	code := g.ind() + fmt.Sprintf("if period <= 0 || ctx.BarIndex < %s {\n", warmupExpr)
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += bodyFn()
	g.indent--
	code += g.ind() + "}\n"
	return code
}

type DynamicSMAEmitter struct{}

func (DynamicSMAEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period-1", func() string {
		code := g.ind() + "sum := 0.0\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("sum += %s.Get(j)\n", sourceAccessor)
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + fmt.Sprintf("%sSeries.Set(sum / float64(period))\n", varName)
		return code
	})
}

type DynamicEMAEmitter struct{}

func (DynamicEMAEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	code := g.ind() + "if period <= 0 {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + "alpha := 2.0 / (float64(period) + 1.0)\n"
	code += g.ind() + fmt.Sprintf("src := %s.Get(0)\n", sourceAccessor)
	code += g.ind() + "if ctx.BarIndex == 0 {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(src)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("prev := %sSeries.Get(1)\n", varName)
	code += g.ind() + "if math.IsNaN(prev) {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(src)\n", varName)
	g.indent--
	code += g.ind() + "} else {\n"
	g.indent++
	code += g.ind() + fmt.Sprintf("%sSeries.Set(alpha*src + (1.0-alpha)*prev)\n", varName)
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	g.indent--
	code += g.ind() + "}\n"
	return code
}

type DynamicRSIEmitter struct{}

func (DynamicRSIEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period", func() string {
		code := g.ind() + "var gains, losses float64\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("change := %s.Get(j) - %s.Get(j+1)\n", sourceAccessor, sourceAccessor)
		code += g.ind() + "if change > 0 {\n"
		g.indent++
		code += g.ind() + "gains += change\n"
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + "losses -= change\n"
		g.indent--
		code += g.ind() + "}\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + "avgGain := gains / float64(period)\n"
		code += g.ind() + "avgLoss := losses / float64(period)\n"
		code += g.ind() + "if avgLoss == 0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(100.0)\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + "rs := avgGain / avgLoss\n"
		code += g.ind() + fmt.Sprintf("%sSeries.Set(100.0 - 100.0/(1.0+rs))\n", varName)
		g.indent--
		code += g.ind() + "}\n"
		return code
	})
}

type DynamicSTDEVEmitter struct{}

func (DynamicSTDEVEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period-1", func() string {
		code := g.ind() + "sum := 0.0\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("sum += %s.Get(j)\n", sourceAccessor)
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + "mean := sum / float64(period)\n"
		code += g.ind() + "variance := 0.0\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("diff := %s.Get(j) - mean\n", sourceAccessor)
		code += g.ind() + "variance += diff * diff\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + fmt.Sprintf("%sSeries.Set(math.Sqrt(variance / float64(period)))\n", varName)
		return code
	})
}

type DynamicHighestEmitter struct{}

func (DynamicHighestEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period-1", func() string {
		code := g.ind() + fmt.Sprintf("maxVal := %s.Get(0)\n", sourceAccessor)
		code += g.ind() + "for j := 1; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("val := %s.Get(j)\n", sourceAccessor)
		code += g.ind() + "if val > maxVal {\n"
		g.indent++
		code += g.ind() + "maxVal = val\n"
		g.indent--
		code += g.ind() + "}\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + fmt.Sprintf("%sSeries.Set(maxVal)\n", varName)
		return code
	})
}

type DynamicLowestEmitter struct{}

func (DynamicLowestEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period-1", func() string {
		code := g.ind() + fmt.Sprintf("minVal := %s.Get(0)\n", sourceAccessor)
		code += g.ind() + "for j := 1; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("val := %s.Get(j)\n", sourceAccessor)
		code += g.ind() + "if val < minVal {\n"
		g.indent++
		code += g.ind() + "minVal = val\n"
		g.indent--
		code += g.ind() + "}\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + fmt.Sprintf("%sSeries.Set(minVal)\n", varName)
		return code
	})
}

type DynamicATREmitter struct{}

func (DynamicATREmitter) EmitCalculation(g *generator, varName, _ string) string {
	return emitWarmupGuard(g, varName, "period", func() string {
		code := g.ind() + "sum := 0.0\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + "high := highSeries.Get(j)\n"
		code += g.ind() + "low := lowSeries.Get(j)\n"
		code += g.ind() + "prevClose := closeSeries.Get(j + 1)\n"
		code += g.ind() + "tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))\n"
		code += g.ind() + "sum += tr\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + fmt.Sprintf("%sSeries.Set(sum / float64(period))\n", varName)
		return code
	})
}

type DynamicCCIEmitter struct{}

func (DynamicCCIEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period-1", func() string {
		code := g.ind() + "sma := 0.0\n"
		code += g.ind() + fmt.Sprintf("for j := 0; j < period; j++ { sma += %s.Get(j) }\n", sourceAccessor)
		code += g.ind() + "sma /= float64(period)\n"
		code += g.ind() + "dev := 0.0\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("v := %s.Get(j)\n", sourceAccessor)
		code += g.ind() + "if v > sma { dev += v - sma } else { dev += sma - v }\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + "dev /= float64(period)\n"
		code += g.ind() + "if dev == 0.0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set((%s.Get(0) - sma) / (0.015 * dev))\n", varName, sourceAccessor)
		g.indent--
		code += g.ind() + "}\n"
		return code
	})
}

type DynamicCOGEmitter struct{}

func (DynamicCOGEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period-1", func() string {
		code := g.ind() + "num, den := 0.0, 0.0\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("v := %s.Get(j)\n", sourceAccessor)
		code += g.ind() + "num += v * float64(j+1)\n"
		code += g.ind() + "den += v\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + "if den == 0.0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(-num / den)\n", varName)
		g.indent--
		code += g.ind() + "}\n"
		return code
	})
}

/* DynamicBBWEmitter captures mult at construction time so it is available during dynamic-period
 * code emission — mult is always compile-time constant (Pine Script simple float) while length
 * may be a runtime series. This avoids altering the shared DynamicPeriodEmitter interface. */
type DynamicBBWEmitter struct{ mult float64 }

func (e DynamicBBWEmitter) EmitCalculation(g *generator, varName, sourceAccessor string) string {
	return emitWarmupGuard(g, varName, "period-1", func() string {
		code := g.ind() + "sma := 0.0\n"
		code += g.ind() + fmt.Sprintf("for j := 0; j < period; j++ { sma += %s.Get(j) }\n", sourceAccessor)
		code += g.ind() + "sma /= float64(period)\n"
		code += g.ind() + "variance := 0.0\n"
		code += g.ind() + "for j := 0; j < period; j++ {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("d := %s.Get(j) - sma\n", sourceAccessor)
		code += g.ind() + "variance += d * d\n"
		g.indent--
		code += g.ind() + "}\n"
		code += g.ind() + "sd := math.Sqrt(variance / float64(period))\n"
		code += g.ind() + "if sma == 0.0 {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(0.0)\n", varName)
		g.indent--
		code += g.ind() + "} else {\n"
		g.indent++
		code += g.ind() + fmt.Sprintf("%sSeries.Set(2.0 * %g * sd / sma)\n", varName, e.mult)
		g.indent--
		code += g.ind() + "}\n"
		return code
	})
}
