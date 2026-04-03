package codegen

import "fmt"

/* Coerces namespace expressions for safe use inside Series.Set(float64) */
type SeriesInitCoercer struct{}

func NewSeriesInitCoercer() *SeriesInitCoercer {
	return &SeriesInitCoercer{}
}

func (c *SeriesInitCoercer) Coerce(code string, goType GoValueType) string {
	switch goType {
	case GoBool:
		return c.boolToFloat64(code)
	case GoString:
		return c.stringFallback(code)
	default:
		return code
	}
}

func (c *SeriesInitCoercer) boolToFloat64(code string) string {
	return fmt.Sprintf("func() float64 { if %s { return 1.0 } else { return 0.0 } }()", code)
}

func (c *SeriesInitCoercer) stringFallback(code string) string {
	return fmt.Sprintf("math.NaN() /* string expression %s cannot be stored in float64 series */", code)
}
