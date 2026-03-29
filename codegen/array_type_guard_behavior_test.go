package codegen

import (
	"strings"
	"testing"
)

func TestArrayOperationTypeGuards_StatisticsOperations(t *testing.T) {
	statisticsOps := []string{
		"array.sum(arr)",
		"array.avg(arr)",
		"array.min(arr)",
		"array.max(arr)",
		"array.median(arr)",
		"array.mode(arr)",
		"array.range(arr)",
		"array.stdev(arr)",
		"array.variance(arr)",
		"array.percentile_linear_interpolation(arr, 50)",
		"array.percentile_nearest_rank(arr, 50)",
		"array.percentrank(arr, 1.5)",
		"array.standardize(arr)",
		"array.abs(arr)",
	}

	for _, op := range statisticsOps {
		t.Run(op, func(t *testing.T) {
			floatSource := "//@version=5\nindicator('test')\narr = array.new_float()\nx = " + op
			_, floatErr := compilePineScript(floatSource)
			if floatErr != nil {
				t.Errorf("Statistics operation %s should succeed on float array: %v", op, floatErr)
			}

			stringSource := "//@version=5\nindicator('test')\narr = array.new_string()\nx = " + op
			_, stringErr := compilePineScript(stringSource)
			if stringErr == nil {
				t.Errorf("Statistics operation %s should fail on string array", op)
			}
			if stringErr != nil && !strings.Contains(stringErr.Error(), "requires numeric array") {
				t.Errorf("Error should mention 'requires numeric array', got: %v", stringErr)
			}
		})
	}
}

func TestArrayOperationTypeGuards_UniversalOperations(t *testing.T) {
	universalOps := []struct {
		op             string
		expectedFloat  string
		expectedString string
	}{
		{"array.push(arr, val)", "NewMutator()", "NewStringMutator()"},
		{"array.pop(arr)", "NewMutator()", "NewStringMutator()"},
		{"array.shift(arr)", "NewMutator()", "NewStringMutator()"},
		{"array.unshift(arr, val)", "NewMutator()", "NewStringMutator()"},
		{"array.clear(arr)", "NewMutator()", "NewStringMutator()"},
		{"array.reverse(arr)", "NewTransformer()", "NewStringTransformers()"},
		{"array.sort(arr)", "NewTransformer()", "NewStringTransformers()"},
		{"array.includes(arr, val)", "NewAccessor()", "NewStringAccessor()"},
		{"array.indexof(arr, val)", "NewAccessor()", "NewStringAccessor()"},
		{"array.join(arr)", "NewFormatters()", "NewStringFormatters()"},
	}

	for _, tc := range universalOps {
		t.Run(tc.op, func(t *testing.T) {
			floatOp := strings.ReplaceAll(tc.op, "val", "1.0")
			floatSource := "//@version=5\nindicator('test')\narr = array.new_float()\nx = " + floatOp
			floatCode, floatErr := compilePineScript(floatSource)
			if floatErr != nil {
				t.Errorf("Operation %s should succeed on float array: %v", tc.op, floatErr)
			}
			if floatErr == nil && !strings.Contains(floatCode, tc.expectedFloat) {
				t.Errorf("Float array operation should use %s", tc.expectedFloat)
			}

			stringOp := strings.ReplaceAll(tc.op, "val", "\"x\"")
			stringSource := "//@version=5\nindicator('test')\narr = array.new_string()\nx = " + stringOp
			stringCode, stringErr := compilePineScript(stringSource)
			if stringErr != nil {
				t.Errorf("Operation %s should succeed on string array: %v", tc.op, stringErr)
			}
			if stringErr == nil && !strings.Contains(stringCode, tc.expectedString) {
				t.Errorf("String array operation should use %s", tc.expectedString)
			}
		})
	}
}

func TestArrayTypeInference_ArrayFromConstructor(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		expectSuffix string
		rejectSuffix string
	}{
		{
			name:         "array.from with numeric literals",
			source:       "//@version=5\nindicator('test')\narr = array.from(1, 2, 3)",
			expectSuffix: "ArraySeries",
			rejectSuffix: "StringArraySeries",
		},
		{
			name:         "array.from allows statistics",
			source:       "//@version=5\nindicator('test')\narr = array.from(1, 2, 3)\nx = array.sum(arr)",
			expectSuffix: "Statistics()",
			rejectSuffix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.source)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}
			if !strings.Contains(code, tt.expectSuffix) {
				t.Errorf("Expected %q in generated code", tt.expectSuffix)
			}
			if tt.rejectSuffix != "" && strings.Contains(code, tt.rejectSuffix) {
				t.Errorf("Should not contain %q in generated code", tt.rejectSuffix)
			}
		})
	}
}

func TestArrayTypeConsistency_MixedDeclarations(t *testing.T) {
	source := `//@version=5
indicator('test')
floats = array.new_float()
strings = array.new_string()
array.push(floats, 1.0)
array.push(strings, "hello")`

	code, err := compilePineScript(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	expectations := []struct {
		pattern     string
		description string
	}{
		{"floatsArraySeries", "float array variable suffix"},
		{"stringsStringArraySeries", "string array variable suffix"},
		{"arrayops.NewMutator().Push(floatsArraySeries,", "float mutator routing"},
		{"arrayops.NewStringMutator().Push(stringsStringArraySeries,", "string mutator routing"},
		{"floatsArraySeries.Next()", "float cursor advancement"},
		{"stringsStringArraySeries.Next()", "string cursor advancement"},
	}

	for _, exp := range expectations {
		if !strings.Contains(code, exp.pattern) {
			t.Errorf("Missing %s: expected %q", exp.description, exp.pattern)
		}
	}
}

func TestArrayElementType_SupportsStatisticsGuard(t *testing.T) {
	tests := []struct {
		elemType ArrayElementType
		supports bool
	}{
		{ArrayElementFloat64, true},
		{ArrayElementString, false},
	}

	for _, tt := range tests {
		t.Run(tt.elemType.GoType(), func(t *testing.T) {
			if tt.elemType.SupportsStatistics() != tt.supports {
				t.Errorf("SupportsStatistics() = %v, want %v", tt.elemType.SupportsStatistics(), tt.supports)
			}

			if tt.supports && tt.elemType.StatisticsConstructor() == "" {
				t.Error("Numeric type should return non-empty StatisticsConstructor()")
			}
		})
	}
}

func TestArrayTypeGuard_CrossTypeOperationRejection(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		shouldFail  bool
		errorSubstr string
	}{
		{
			name:        "string array with numeric operation",
			source:      "//@version=5\nindicator('test')\narr = array.new_string()\nx = array.sum(arr)",
			shouldFail:  true,
			errorSubstr: "requires numeric array",
		},
		{
			name:        "float array with join operation",
			source:      "//@version=5\nindicator('test')\narr = array.new_float()\nx = array.join(arr)",
			shouldFail:  false,
			errorSubstr: "",
		},
		{
			name:        "string array with variance",
			source:      "//@version=5\nindicator('test')\narr = array.new_string()\nx = array.variance(arr)",
			shouldFail:  true,
			errorSubstr: "requires numeric array",
		},
		{
			name:        "string array with push",
			source:      "//@version=5\nindicator('test')\narr = array.new_string()\narray.push(arr, \"test\")",
			shouldFail:  false,
			errorSubstr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compilePineScript(tt.source)
			if tt.shouldFail && err == nil {
				t.Error("Expected compilation to fail")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected compilation to succeed: %v", err)
			}
			if tt.shouldFail && err != nil && tt.errorSubstr != "" {
				if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Error should contain %q, got: %v", tt.errorSubstr, err)
				}
			}
		})
	}
}

func TestArrayTypeGuard_GetOperationRoutingDifference(t *testing.T) {
	floatSource := "//@version=5\nindicator('test')\narr = array.new_float()\nx = array.get(arr, 0)"
	floatCode, err := compilePineScript(floatSource)
	if err != nil {
		t.Fatalf("Float array get failed: %v", err)
	}
	if !strings.Contains(floatCode, "arrArraySeries.Elem") {
		t.Error("Float array get should use .Elem on ArraySeries")
	}

	stringSource := "//@version=5\nindicator('test')\narr = array.new_string()\nx = array.get(arr, 0)"
	stringCode, err := compilePineScript(stringSource)
	if err != nil {
		t.Fatalf("String array get failed: %v", err)
	}
	if !strings.Contains(stringCode, "arrStringArraySeries.Elem") {
		t.Error("String array get should use .Elem on StringArraySeries")
	}
}
