package codegen

type SelfReferencingIIFEGenerator interface {
	GenerateWithSelfReference(accessor AccessGenerator, targetSeriesVar string) string
}

type FixnanIIFEGenerator struct{}

func (g *FixnanIIFEGenerator) GenerateWithSelfReference(accessor AccessGenerator, targetSeriesVar string) string {
	var preamble string
	if tempAccessor, ok := accessor.(*FixnanCallExpressionAccessor); ok {
		preamble = tempAccessor.GetPreamble()
	}

	body := "val := " + accessor.GenerateLoopValueAccess("0") + "; "
	body += "if math.IsNaN(val) { return 0.0 }; "
	body += "return val"

	if preamble != "" {
		return preamble + "func() float64 { " + body + " }()"
	}
	return "func() float64 { " + body + " }()"
}

type FixnanCallExpressionAccessor struct {
	tempVarName string
	tempVarCode string
}

func (a *FixnanCallExpressionAccessor) GenerateLoopValueAccess(loopVar string) string {
	return a.tempVarName
}

func (a *FixnanCallExpressionAccessor) GenerateInitialValueAccess(period int) string {
	return a.tempVarName
}

func (a *FixnanCallExpressionAccessor) GetPreamble() string {
	return a.tempVarCode
}
