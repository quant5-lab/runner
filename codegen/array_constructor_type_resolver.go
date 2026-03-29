package codegen

type ArrayConstructorTypeResolver struct {
	classifier *ArrayConstructorClassifier
}

func NewArrayConstructorTypeResolver() *ArrayConstructorTypeResolver {
	return &ArrayConstructorTypeResolver{
		classifier: NewArrayConstructorClassifier(),
	}
}

func (r *ArrayConstructorTypeResolver) ResolveVariableType(funcName string) (ArrayElementType, bool) {
	if r.classifier.IsStringArrayConstructor(funcName) {
		return ArrayElementString, true
	}

	if r.classifier.IsNumericArrayConstructor(funcName) {
		return ArrayElementFloat64, true
	}

	if r.classifier.IsDrawingArrayConstructor(funcName) {
		return ArrayElementFloat64, true
	}

	if funcName == "array.from" {
		return ArrayElementFloat64, true
	}

	if funcName == "ta.pivot_point_levels" || funcName == "pivot_point_levels" {
		return ArrayElementFloat64, true
	}

	return 0, false
}
