package codegen

type ColorFunctionSpec struct {
	GoName     string
	ReturnType string
}

var colorFunctionSpecs = map[string]ColorFunctionSpec{
	"color.new":           {GoName: "visual.PineColorNew", ReturnType: "string"},
	"color.rgb":           {GoName: "visual.PineColorRGB", ReturnType: "string"},
	"color.from_gradient": {GoName: "visual.PineColorFromGradient", ReturnType: "string"},
	"color.r":             {GoName: "visual.PineColorR", ReturnType: "float64"},
	"color.g":             {GoName: "visual.PineColorG", ReturnType: "float64"},
	"color.b":             {GoName: "visual.PineColorB", ReturnType: "float64"},
	"color.t":             {GoName: "visual.PineColorT", ReturnType: "float64"},
}

func IsColorFunction(name string) bool {
	_, ok := colorFunctionSpecs[name]
	return ok
}

func ColorFunctionReturnType(name string) string {
	if spec, ok := colorFunctionSpecs[name]; ok {
		return spec.ReturnType
	}
	return ""
}

func ColorFunctionGoName(name string) string {
	if spec, ok := colorFunctionSpecs[name]; ok {
		return spec.GoName
	}
	return ""
}
