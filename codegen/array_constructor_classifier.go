package codegen

type ArrayConstructorClassifier struct{}

func NewArrayConstructorClassifier() *ArrayConstructorClassifier {
	return &ArrayConstructorClassifier{}
}

func (c *ArrayConstructorClassifier) IsArrayConstructor(funcName string) bool {
	return c.IsNumericArrayConstructor(funcName) ||
		c.IsStringArrayConstructor(funcName) ||
		c.IsDrawingArrayConstructor(funcName) ||
		funcName == "array.from"
}

func (c *ArrayConstructorClassifier) IsStringArrayConstructor(funcName string) bool {
	return funcName == "array.new_string"
}

func (c *ArrayConstructorClassifier) IsNumericArrayConstructor(funcName string) bool {
	switch funcName {
	case "array.new_float", "array.new_int", "array.new_bool", "array.new_color":
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

func (c *ArrayConstructorClassifier) IsDrawingArrayConstructor(funcName string) bool {
	switch funcName {
	case "array.new_label", "array.new_line", "array.new_box",
		"array.new_table", "array.new_linefill":
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
		"array.covariance", "array.standardize", "array.abs",
		"array.binary_search", "array.binary_search_leftmost",
		"array.binary_search_rightmost", "array.every", "array.some",
		"array.join":
		return true
	}
	return false
}
