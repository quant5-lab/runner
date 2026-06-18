package codegen

import (
	"fmt"
	"strings"
)

// mapperInitBlock selects the bar-mapping algorithm from four cases based on the
// secondary timeframe relative to the base and whether it is intraday or calendar-scale:
//
//	secTF < baseTF                          → BuildMappingForUpscaling
//	secTF == baseTF, variable bar count     → BuildMappingFromTransform
//	secTF == baseTF, 1:1 bars              → BuildIdentityMapping
//	secTF > baseTF, intraday (< 86400 s)   → BuildMappingByTimestamp
//	secTF > baseTF, calendar-scale (≥ 1 D) → BuildMappingWithDateFilter
//
// The intraday vs calendar-scale split matters because multiple intraday coarser
// bars can share the same calendar date (e.g. four 4h bars per trading day).
// Date-string comparison is ambiguous in that case and conflates them into one
// range.  Timestamp comparison is unambiguous and correctly handles session-
// anchored boundaries that may straddle midnight.  Calendar-scale coarser bars
// (daily, weekly, monthly) each have a unique date, so the existing date-filter
// path is kept unchanged for backward compatibility with existing goldens.
func mapperInitBlock(varName string, variableBarCount bool) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t%s_mapper := request.NewSecurityBarMapper()\n", varName))
	b.WriteString("\tif secTimeframeSeconds < baseTimeframeSeconds {\n")
	b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingForUpscaling(%s_ctx.Data, ctx.Data, ctx.Timezone)\n", varName, varName))
	b.WriteString("\t} else if secTimeframeSeconds == baseTimeframeSeconds {\n")
	if variableBarCount {
		b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingFromTransform(%s_transformResult.MainToSynthetic)\n", varName, varName))
	} else {
		b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildIdentityMapping(len(ctx.Data))\n", varName))
	}
	b.WriteString("\t} else if secTimeframeSeconds < 86400 {\n")
	b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingByTimestamp(%s_ctx.Data, ctx.Data)\n", varName, varName))
	b.WriteString("\t} else {\n")
	b.WriteString(fmt.Sprintf("\t\t%s_mapper.BuildMappingWithDateFilter(%s_ctx.Data, ctx.Data, baseDateRange, ctx.Timezone)\n", varName, varName))
	b.WriteString("\t}\n")
	return b.String()
}
