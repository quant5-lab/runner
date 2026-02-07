package codegen

import "strings"

type MathFunctionSpec struct {
	PineName        string
	GoFunc          string
	MinArgs         int
	MaxArgs         int
	GeneratorMethod string
}

type MathFunctionRegistry struct {
	functions map[string]*MathFunctionSpec
}

func NewMathFunctionRegistry() *MathFunctionRegistry {
	r := &MathFunctionRegistry{
		functions: make(map[string]*MathFunctionSpec),
	}
	r.registerAll()
	return r
}

func (r *MathFunctionRegistry) Lookup(name string) (*MathFunctionSpec, bool) {
	normalized := strings.ToLower(name)
	if !strings.HasPrefix(normalized, "math.") {
		normalized = "math." + normalized
	}
	spec, ok := r.functions[normalized]
	return spec, ok
}

func (r *MathFunctionRegistry) register(pineName, goFunc string, minArgs, maxArgs int, method string) {
	r.functions[pineName] = &MathFunctionSpec{
		PineName:        pineName,
		GoFunc:          goFunc,
		MinArgs:         minArgs,
		MaxArgs:         maxArgs,
		GeneratorMethod: method,
	}
}

func (r *MathFunctionRegistry) registerAll() {
	r.register("math.abs", "math.Abs", 1, 1, "unary")
	r.register("math.acos", "math.Acos", 1, 1, "unary")
	r.register("math.asin", "math.Asin", 1, 1, "unary")
	r.register("math.atan", "math.Atan", 1, 1, "unary")
	r.register("math.ceil", "math.Ceil", 1, 1, "unary")
	r.register("math.cos", "math.Cos", 1, 1, "unary")
	r.register("math.exp", "math.Exp", 1, 1, "unary")
	r.register("math.floor", "math.Floor", 1, 1, "unary")
	r.register("math.log", "math.Log", 1, 1, "unary")
	r.register("math.log10", "math.Log10", 1, 1, "unary")
	r.register("math.round", "math.Round", 1, 1, "unary")
	r.register("math.sin", "math.Sin", 1, 1, "unary")
	r.register("math.sqrt", "math.Sqrt", 1, 1, "unary")
	r.register("math.tan", "math.Tan", 1, 1, "unary")
	r.register("math.max", "math.Max", 2, 2, "binary")
	r.register("math.min", "math.Min", 2, 2, "binary")
	r.register("math.pow", "math.Pow", 2, 2, "binary")
	r.register("math.sign", "", 1, 1, "sign")
	r.register("math.todegrees", "", 1, 1, "todegrees")
	r.register("math.toradians", "", 1, 1, "toradians")
	r.register("math.avg", "", 1, -1, "avg")
	r.register("math.random", "", 0, 3, "random")
	r.register("math.round_to_mintick", "", 1, 1, "round_to_mintick")
}
