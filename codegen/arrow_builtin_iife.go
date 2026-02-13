package codegen

import "fmt"

func boundsCheckedArrowIIFE(indexCode, innerBody string) string {
	return fmt.Sprintf(
		"func() float64 { barIdx := ctx.BarIndex-int(%s); if barIdx >= 0 && barIdx < len(ctx.Data) { %s }; return math.NaN() }()",
		indexCode, innerBody,
	)
}

func OHLCVFieldArrowIIFE(fieldName, indexCode string) string {
	return boundsCheckedArrowIIFE(indexCode, fmt.Sprintf("return ctx.Data[barIdx].%s", fieldName))
}

func TimeArrowIIFE(indexCode string) string {
	return boundsCheckedArrowIIFE(indexCode, "return float64(ctx.Data[barIdx].Time * 1000)")
}

func CalendarFieldArrowIIFE(arrowExpression, indexCode string) string {
	innerBody := fmt.Sprintf(
		"tz, _ := time.LoadLocation(ctx.Timezone); barTime := time.Unix(ctx.Data[barIdx].Time, 0).In(tz); return %s",
		arrowExpression,
	)
	return boundsCheckedArrowIIFE(indexCode, innerBody)
}

func BarIndexArrowIIFE(indexCode string) string {
	return boundsCheckedArrowIIFE(indexCode, "return float64(barIdx)")
}
