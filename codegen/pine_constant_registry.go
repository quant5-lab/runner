package codegen

type PineConstantRegistry struct {
	constants map[string]ConstantValue
}

func NewPineConstantRegistry() *PineConstantRegistry {
	cr := &PineConstantRegistry{
		constants: make(map[string]ConstantValue),
	}
	cr.registerPineScriptConstants()
	return cr
}

func (cr *PineConstantRegistry) Get(key string) (ConstantValue, bool) {
	val, exists := cr.constants[key]
	return val, exists
}

func (cr *PineConstantRegistry) register(key string, value ConstantValue) {
	cr.constants[key] = value
}

func (cr *PineConstantRegistry) registerPineScriptConstants() {
	cr.registerBarmergeConstants()
	cr.registerStrategyConstants()
	cr.registerColorConstants()
	cr.registerPlotConstants()
}

func (cr *PineConstantRegistry) registerBarmergeConstants() {
	cr.register("barmerge.lookahead_on", NewBoolConstant(true))
	cr.register("barmerge.lookahead_off", NewBoolConstant(false))
	cr.register("barmerge.gaps_on", NewBoolConstant(true))
	cr.register("barmerge.gaps_off", NewBoolConstant(false))
}

func (cr *PineConstantRegistry) registerStrategyConstants() {
	cr.register("strategy.long", NewIntConstant(1))
	cr.register("strategy.short", NewIntConstant(-1))
	cr.register("strategy.cash", NewStringConstant("cash"))
	cr.register("strategy.percent_of_equity", NewStringConstant("percent_of_equity"))
	cr.register("strategy.fixed", NewStringConstant("fixed"))
}

func (cr *PineConstantRegistry) registerColorConstants() {
	cr.register("color.aqua", NewStringConstant("#00BCD4"))
	cr.register("color.black", NewStringConstant("#363A45"))
	cr.register("color.blue", NewStringConstant("#2962FF"))
	cr.register("color.fuchsia", NewStringConstant("#E040FB"))
	cr.register("color.gray", NewStringConstant("#787B86"))
	cr.register("color.green", NewStringConstant("#4CAF50"))
	cr.register("color.lime", NewStringConstant("#00E676"))
	cr.register("color.maroon", NewStringConstant("#880E4F"))
	cr.register("color.navy", NewStringConstant("#311B92"))
	cr.register("color.olive", NewStringConstant("#808000"))
	cr.register("color.orange", NewStringConstant("#FF9800"))
	cr.register("color.purple", NewStringConstant("#9C27B0"))
	cr.register("color.red", NewStringConstant("#FF5252"))
	cr.register("color.silver", NewStringConstant("#B2B5BE"))
	cr.register("color.teal", NewStringConstant("#00897B"))
	cr.register("color.white", NewStringConstant("#FFFFFF"))
	cr.register("color.yellow", NewStringConstant("#FFEB3B"))
}

func (cr *PineConstantRegistry) registerPlotConstants() {
	cr.register("plot.style_line", NewStringConstant("line"))
	cr.register("plot.style_linebr", NewStringConstant("linebr"))
	cr.register("plot.style_stepline", NewStringConstant("stepline"))
	cr.register("plot.style_histogram", NewStringConstant("histogram"))
	cr.register("plot.style_cross", NewStringConstant("cross"))
	cr.register("plot.style_area", NewStringConstant("area"))
	cr.register("plot.style_columns", NewStringConstant("columns"))
	cr.register("plot.style_circles", NewStringConstant("circles"))
}

func (cr *PineConstantRegistry) IsColorName(name string) bool {
	_, exists := cr.constants["color."+name]
	return exists
}

func (cr *PineConstantRegistry) GetColorHex(name string) (string, bool) {
	val, exists := cr.constants["color."+name]
	if !exists {
		return "", false
	}
	return val.AsString()
}
