package codegen

type SelfReferencingIIFEGenerator interface {
	GenerateWithSelfReference(accessor AccessGenerator, targetSeriesVar string) string
}

type FixnanIIFEGenerator struct{}

func (g *FixnanIIFEGenerator) GenerateWithSelfReference(accessor AccessGenerator, targetSeriesVar string) string {
	body := "val := " + accessor.GenerateLoopValueAccess("0") + "; "
	body += "if math.IsNaN(val) { return 0.0 }; "
	body += "return val"

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
