package codegen

import (
	"fmt"
	"strings"
)

// fetchDataBlock emits Go source that populates four per-feed variables for one
// security() call: varName_data, varName_tz, varName_refsession, varName_anchor.
//
// Two acquisition paths are generated inside the block:
//
// Normal path (fixture present): the datafetcher loads the fixture, market
// normalizes the bars, and the secondary session anchor is derived from the
// secondary symbol's own profile and bar data — so a NYSE secondary gets the
// NYSE anchor, a MOEX secondary gets the MOEX anchor, regardless of what the
// primary symbol is.
//
// Aggregate fallback (fixture absent, same symbol, coarser timeframe): the
// primary bars are collapsed with context.AggregateToCoarserPeriod using the
// primary session anchor (ctx.PeriodAnchor), which was already derived from
// observed first-bar open times.  Timezone and reference session are inherited
// from the primary context because the symbol is the same.
//
// Any other missing-fixture case (foreign symbol, or finer/equal timeframe)
// terminates the binary — data fabrication for a foreign symbol would silently
// corrupt strategy output.
func fetchDataBlock(varName, symbolCode, timeframeCode string, warmupBars int) string {
	var b strings.Builder
	emitLimitBlock(&b, varName, warmupBars)
	emitDataVarDeclarations(&b, varName)
	emitFetchWithFallback(&b, varName, symbolCode, timeframeCode)
	emitTrimBlock(&b, varName)
	return b.String()
}

// emitLimitBlock caps secondary bar count to the strategy's look-back window.
func emitLimitBlock(b *strings.Builder, varName string, warmupBars int) {
	b.WriteString(fmt.Sprintf("\t%s_limit := len(ctx.Data)\n", varName))
	b.WriteString("\tif secTimeframeSeconds != baseTimeframeSeconds && len(ctx.Data) > 0 {\n")
	b.WriteString("\t\tfirstBarTime := ctx.Data[0].Time\n")
	b.WriteString("\t\tlastBarTime := ctx.Data[len(ctx.Data)-1].Time\n")
	b.WriteString("\t\ttimeSpanSeconds := lastBarTime - firstBarTime\n")
	b.WriteString("\t\tbaseSecurityBars := int(timeSpanSeconds/secTimeframeSeconds) + 1\n")
	b.WriteString(fmt.Sprintf("\t\tdetectedWarmup := %d\n", warmupBars))
	b.WriteString("\t\tfixedMinimumWarmup := 500\n")
	b.WriteString("\t\trequiredWarmup := fixedMinimumWarmup\n")
	b.WriteString("\t\tif detectedWarmup > requiredWarmup {\n")
	b.WriteString("\t\t\trequiredWarmup = detectedWarmup\n")
	b.WriteString("\t\t}\n")
	b.WriteString(fmt.Sprintf("\t\t%s_limit = baseSecurityBars + requiredWarmup\n", varName))
	b.WriteString("\t}\n")
}

// emitDataVarDeclarations pre-initialises the output variables from the primary
// context so the aggregate-fallback path needs no extra assignments.
func emitDataVarDeclarations(b *strings.Builder, varName string) {
	b.WriteString(fmt.Sprintf("\tvar %s_data []context.OHLCV\n", varName))
	b.WriteString(fmt.Sprintf("\t%s_tz := ctx.Timezone\n", varName))
	b.WriteString(fmt.Sprintf("\t%s_refsession := ctx.ReferenceSession\n", varName))
	b.WriteString(fmt.Sprintf("\t%s_anchor := ctx.PeriodAnchor\n", varName))
}

func emitFetchWithFallback(b *strings.Builder, varName, symbolCode, timeframeCode string) {
	b.WriteString(fmt.Sprintf("\t%s_marketData, %s_err := fetcher.FetchWithMetadata(%s, %s, 0)\n", varName, varName, symbolCode, timeframeCode))
	b.WriteString(fmt.Sprintf("\tif %s_err != nil {\n", varName))
	emitAggregateFallbackGuard(b, varName, symbolCode, timeframeCode)
	b.WriteString("\t} else {\n")
	emitNormalizePath(b, varName, symbolCode, timeframeCode)
	b.WriteString("\t}\n")
}

// emitAggregateFallbackGuard guards against data fabrication: a same-symbol coarser
// request aggregates in-process; any other missing fixture terminates the binary.
func emitAggregateFallbackGuard(b *strings.Builder, varName, symbolCode, timeframeCode string) {
	b.WriteString(fmt.Sprintf("\t\tif secTimeframeSeconds > baseTimeframeSeconds && %s == ctx.Symbol {\n", symbolCode))
	b.WriteString(fmt.Sprintf("\t\t\t%s_data = context.AggregateToCoarserPeriod(ctx.Data, %s, ctx.PeriodAnchor)\n", varName, timeframeCode))
	b.WriteString("\t\t} else {\n")
	b.WriteString(fmt.Sprintf("\t\t\tfmt.Fprintf(os.Stderr, \"Failed to fetch %%s:%%s: %%v\\n\", %s, %s, %s_err)\n", symbolCode, timeframeCode, varName))
	b.WriteString("\t\t\tos.Exit(1)\n")
	b.WriteString("\t\t}\n")
}

// emitNormalizePath resolves the secondary session anchor from the secondary
// symbol's own exchange, so each exchange (NYSE, MOEX, …) uses its own
// session-open offset independently of the primary symbol's exchange.
func emitNormalizePath(b *strings.Builder, varName, symbolCode, timeframeCode string) {
	b.WriteString(fmt.Sprintf("\t\t%s_metadata := market.CompleteSecondaryMetadata(%s, %s_marketData.SourceMetadata, market.SecondaryContextDefaults{Timezone: ctx.Timezone, ReferenceSession: ctx.ReferenceSession})\n", varName, symbolCode, varName))
	b.WriteString(fmt.Sprintf("\t\t%s_normalized, %s_profile, %s_normErr := market.NormalizeBarsWithMetadataE(%s, %s, %s_metadata, %s_marketData.Bars)\n",
		varName, varName, varName, symbolCode, timeframeCode, varName, varName))
	b.WriteString(fmt.Sprintf("\t\tif %s_normErr != nil {\n", varName))
	b.WriteString(fmt.Sprintf("\t\t\tfmt.Fprintf(os.Stderr, \"Invalid market metadata for %%s:%%s: %%v\\n\", %s, %s, %s_normErr)\n", symbolCode, timeframeCode, varName))
	b.WriteString("\t\t\tos.Exit(1)\n")
	b.WriteString("\t\t}\n")
	b.WriteString(fmt.Sprintf("\t\t%s_data = %s_normalized\n", varName, varName))
	b.WriteString(fmt.Sprintf("\t\t%s_tz = %s_profile.Timezone\n", varName, varName))
	b.WriteString(fmt.Sprintf("\t\t%s_refsession = string(%s_profile.ReferenceSession)\n", varName, varName))
	b.WriteString(fmt.Sprintf("\t\t%s_anchor = market.DeriveSessionAnchor(%s_profile, %s_normalized)\n", varName, varName, varName))
}

// emitTrimBlock trims varName_data to the pre-computed limit so that the secondary
// context does not carry excess warmup bars into the bar-evaluation loop.
func emitTrimBlock(b *strings.Builder, varName string) {
	b.WriteString(fmt.Sprintf("\tif %s_limit > 0 && %s_limit < len(%s_data) {\n", varName, varName, varName))
	b.WriteString(fmt.Sprintf("\t\t%s_data = %s_data[len(%s_data)-%s_limit:]\n", varName, varName, varName, varName))
	b.WriteString("\t}\n")
}
