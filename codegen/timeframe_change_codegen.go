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
		"currAligned := context.AlignTimestampToPeriod(ctx.Data[ctx.BarIndex].Time, tf); "+
		"if ctx.BarIndex == 0 { return %s }; "+
		"prevAligned := context.AlignTimestampToPeriod(ctx.Data[ctx.BarIndex-1].Time, tf); "+
		"%s "+
		"}())", style.returnType, tfExpr, style.firstBar, style.comparison)
}
