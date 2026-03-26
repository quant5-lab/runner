package codegen

type ArrayConstructorClassifier struct{}

func NewArrayConstructorClassifier() *ArrayConstructorClassifier {
	return &ArrayConstructorClassifier{}
}

func (c *ArrayConstructorClassifier) IsArrayConstructor(funcName string) bool {
	switch funcName {
	case "array.new_float", "array.new_int", "array.new_bool", "array.from":
		return true
	}
	return false
}

func (c *ArrayConstructorClassifier) IsMutatingMethod(funcName string) bool {
	switch funcName {
	case "array.push", "array.pop", "array.shift", "array.unshift",
		"array.set", "array.insert", "array.remove", "array.clear",
		"array.fill", "array.reverse", "array.sort", "array.concat":
		return true
	}
	return false
}

func (c *ArrayConstructorClassifier) IsReadOnlyMethod(funcName string) bool {
	switch funcName {
	case "array.get", "array.size", "array.first", "array.last",
		"array.includes", "array.indexof", "array.lastindexof",
		"array.slice", "array.copy", "array.sort_indices",
		"array.sum", "array.avg", "array.min", "array.max",
		"array.median", "array.mode", "array.stdev", "array.variance",
		"array.range", "array.percentile_linear_interpolation",
		"array.percentile_nearest_rank", "array.percentrank",
		"array.covariance", "array.standardize", "array.abs":
		return true
	}
	return false
}
