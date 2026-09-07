package codegen

import "fmt"

type timeframeChangeStyle struct {
	returnType string
	firstBar   string
	comparison string
}

var (
	timeframeChangeFloat = timeframeChangeStyle{
		returnType: "float64",
		firstBar:   "1",
		comparison: "if currAligned != prevAligned { return 1 }; return 0",
	}

	timeframeChangeBool = timeframeChangeStyle{
		returnType: "bool",
		firstBar:   "true",
		comparison: "return currAligned != prevAligned",
	}
)

func buildTimeframeChangeIIFE(tfExpr string, style timeframeChangeStyle) string {
	return fmt.Sprintf("(func() %s { "+
		"tf := %s; "+
		"currAligned := context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex].Time, tf, ctx.PeriodAnchor); "+
		"if ctx.BarIndex == 0 { return %s }; "+
		"prevAligned := context.AlignTimestampToPeriodWithAnchor(ctx.Data[ctx.BarIndex-1].Time, tf, ctx.PeriodAnchor); "+
		"%s "+
		"}())", style.returnType, tfExpr, style.firstBar, style.comparison)
}
