package codegen

import "testing"

func TestArrayConstructorClassifier_IsArrayConstructor(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"array.new_float", "array.new_float", true},
		{"array.new_int", "array.new_int", true},
		{"array.new_bool", "array.new_bool", true},
		{"array.new_color", "array.new_color", true},
		{"array.from", "array.from", true},
		{"array.new_label", "array.new_label", true},
		{"array.new_line", "array.new_line", true},
		{"array.new_box", "array.new_box", true},
		{"array.new_table", "array.new_table", true},
		{"array.new_linefill", "array.new_linefill", true},

		{"array.new_string not supported yet", "array.new_string", false},
		{"array.push is mutator not constructor", "array.push", false},
		{"array.get is reader not constructor", "array.get", false},
		{"array.size is reader not constructor", "array.size", false},
		{"ta.sma is not array namespace", "ta.sma", false},
		{"empty string", "", false},
		{"partial match new_float", "new_float", false},
		{"partial match array", "array", false},
		{"case sensitive array.New_Float", "array.New_Float", false},
		{"typo array.new_floats", "array.new_floats", false},
		{"unknown array.new_map", "array.new_map", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsArrayConstructor(tt.funcName)
			if got != tt.want {
				t.Errorf("IsArrayConstructor(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestArrayConstructorClassifier_IsStringArrayConstructor(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"array.new_string not supported yet", "array.new_string", false},
		{"array.new_float is numeric not string", "array.new_float", false},
		{"array.new_int is numeric not string", "array.new_int", false},
		{"array.new_bool is numeric not string", "array.new_bool", false},
		{"array.new_color is numeric not string", "array.new_color", false},
		{"array.from is polymorphic not string-specific", "array.from", false},
		{"empty string", "", false},
		{"partial match new_string", "new_string", false},
		{"case sensitive array.new_String", "array.new_String", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsStringArrayConstructor(tt.funcName)
			if got != tt.want {
				t.Errorf("IsStringArrayConstructor(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestArrayConstructorClassifier_IsNumericArrayConstructor(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"array.new_float", "array.new_float", true},
		{"array.new_int", "array.new_int", true},
		{"array.new_bool", "array.new_bool", true},
		{"array.new_color", "array.new_color", true},

		{"array.new_string is string not numeric", "array.new_string", false},
		{"array.from is polymorphic not numeric-specific", "array.from", false},
		{"array.push is mutator not constructor", "array.push", false},
		{"array.get is reader not constructor", "array.get", false},
		{"empty string", "", false},
		{"unknown array.new_double", "array.new_double", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsNumericArrayConstructor(tt.funcName)
			if got != tt.want {
				t.Errorf("IsNumericArrayConstructor(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestArrayConstructorClassifier_IsMutatingMethod(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"array.push", "array.push", true},
		{"array.pop", "array.pop", true},
		{"array.shift", "array.shift", true},
		{"array.unshift", "array.unshift", true},
		{"array.set", "array.set", true},
		{"array.insert", "array.insert", true},
		{"array.remove", "array.remove", true},
		{"array.clear", "array.clear", true},
		{"array.fill", "array.fill", true},
		{"array.reverse", "array.reverse", true},
		{"array.sort", "array.sort", true},
		{"array.concat", "array.concat", true},

		{"array.get is reader not mutator", "array.get", false},
		{"array.size is reader not mutator", "array.size", false},
		{"array.new_float is constructor not mutator", "array.new_float", false},
		{"array.from is constructor not mutator", "array.from", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsMutatingMethod(tt.funcName)
			if got != tt.want {
				t.Errorf("IsMutatingMethod(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestArrayConstructorClassifier_IsReadOnlyMethod(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"array.get", "array.get", true},
		{"array.size", "array.size", true},
		{"array.first", "array.first", true},
		{"array.last", "array.last", true},
		{"array.includes", "array.includes", true},
		{"array.indexof", "array.indexof", true},
		{"array.lastindexof", "array.lastindexof", true},
		{"array.slice", "array.slice", true},
		{"array.copy", "array.copy", true},
		{"array.sum", "array.sum", true},
		{"array.avg", "array.avg", true},
		{"array.min", "array.min", true},
		{"array.max", "array.max", true},
		{"array.median", "array.median", true},
		{"array.mode", "array.mode", true},
		{"array.stdev", "array.stdev", true},
		{"array.variance", "array.variance", true},
		{"array.range", "array.range", true},
		{"array.join", "array.join", true},
		{"array.binary_search", "array.binary_search", true},
		{"array.every", "array.every", true},
		{"array.some", "array.some", true},

		{"array.push is mutator not reader", "array.push", false},
		{"array.pop is mutator not reader", "array.pop", false},
		{"array.new_float is constructor not reader", "array.new_float", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsReadOnlyMethod(tt.funcName)
			if got != tt.want {
				t.Errorf("IsReadOnlyMethod(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestArrayConstructorClassifier_IsDrawingArrayConstructor(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{"array.new_label", "array.new_label", true},
		{"array.new_line", "array.new_line", true},
		{"array.new_box", "array.new_box", true},
		{"array.new_table", "array.new_table", true},
		{"array.new_linefill", "array.new_linefill", true},

		{"array.new_float is numeric not drawing", "array.new_float", false},
		{"array.from is polymorphic not drawing", "array.from", false},
		{"typo array.new_labels", "array.new_labels", false},
		{"unknown array.new_polyline", "array.new_polyline", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifier.IsDrawingArrayConstructor(tt.funcName)
			if got != tt.want {
				t.Errorf("IsDrawingArrayConstructor(%q) = %v, want %v", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestArrayConstructorClassifier_MethodCategoryMutualExclusivity(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	allMethods := []string{
		"array.new_float", "array.new_int", "array.new_bool", "array.new_color", "array.from",
		"array.new_label", "array.new_line", "array.new_box", "array.new_table", "array.new_linefill",
		"array.push", "array.pop", "array.shift", "array.unshift", "array.set", "array.insert", "array.remove", "array.clear", "array.fill", "array.reverse", "array.sort", "array.concat",
		"array.get", "array.size", "array.first", "array.last", "array.includes", "array.indexof", "array.slice", "array.sum", "array.join",
	}

	for _, method := range allMethods {
		isConstructor := classifier.IsArrayConstructor(method)
		isMutator := classifier.IsMutatingMethod(method)
		isReader := classifier.IsReadOnlyMethod(method)

		categories := 0
		if isConstructor {
			categories++
		}
		if isMutator {
			categories++
		}
		if isReader {
			categories++
		}

		if categories > 1 {
			t.Errorf("%q classified in multiple categories: constructor=%v mutator=%v reader=%v",
				method, isConstructor, isMutator, isReader)
		}
		if categories == 0 {
			t.Errorf("%q not classified in any category", method)
		}
	}
}

func TestArrayConstructorClassifier_TypeSpecificConstructorMutualExclusivity(t *testing.T) {
	classifier := NewArrayConstructorClassifier()

	numericConstructors := []string{
		"array.new_float", "array.new_int", "array.new_bool", "array.new_color",
	}

	for _, method := range numericConstructors {
		isNumeric := classifier.IsNumericArrayConstructor(method)
		isString := classifier.IsStringArrayConstructor(method)
		isDrawing := classifier.IsDrawingArrayConstructor(method)

		if isNumeric && isString {
			t.Errorf("%q classified as both numeric and string constructor", method)
		}
		if isNumeric && isDrawing {
			t.Errorf("%q classified as both numeric and drawing constructor", method)
		}
		if !isNumeric && !isString && !isDrawing {
			t.Errorf("%q classified as neither numeric, string, nor drawing constructor", method)
		}
	}

	drawingConstructors := []string{
		"array.new_label", "array.new_line", "array.new_box", "array.new_table", "array.new_linefill",
	}

	for _, method := range drawingConstructors {
		isNumeric := classifier.IsNumericArrayConstructor(method)
		isString := classifier.IsStringArrayConstructor(method)
		isDrawing := classifier.IsDrawingArrayConstructor(method)

		if isDrawing && isNumeric {
			t.Errorf("%q classified as both drawing and numeric constructor", method)
		}
		if isDrawing && isString {
			t.Errorf("%q classified as both drawing and string constructor", method)
		}
		if !isDrawing {
			t.Errorf("%q should be classified as drawing constructor", method)
		}
	}

	polymorphicConstructors := []string{"array.from"}
	for _, method := range polymorphicConstructors {
		if classifier.IsNumericArrayConstructor(method) {
			t.Errorf("%q should not be classified as numeric-specific (it's polymorphic)", method)
		}
		if classifier.IsStringArrayConstructor(method) {
			t.Errorf("%q should not be classified as string-specific (it's polymorphic)", method)
		}
		if classifier.IsDrawingArrayConstructor(method) {
			t.Errorf("%q should not be classified as drawing-specific (it's polymorphic)", method)
		}
		if !classifier.IsArrayConstructor(method) {
			t.Errorf("%q should still be recognized as array constructor", method)
		}
	}
}
