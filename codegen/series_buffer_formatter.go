package codegen

import "fmt"

// formatSeriesGet generates code for reading from a series buffer
func formatSeriesGet(varName string, offset int) string {
	return fmt.Sprintf("%sSeries.Get(%d)", varName, offset)
}

// formatSeriesSet generates code for writing to a series buffer
func formatSeriesSet(varName string, value string) string {
	return fmt.Sprintf("%sSeries.Set(%s)", varName, value)
}

// formatArrowSeriesGet generates code for reading from arrow context series
func formatArrowSeriesGet(varName string, offset int) string {
	return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(%d)", varName, offset)
}

// formatArrowSeriesSet generates code for writing to arrow context series
func formatArrowSeriesSet(varName string, value string) string {
	return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Set(%s)", varName, value)
}
