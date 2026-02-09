package codegen

import (
	"strings"
	"testing"
)

func TestAppendWithBareAlias_NamespacedFunction(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
	}

	signatures = appendWithBareAlias(signatures, "ta.sma", "close", overloads)

	if len(signatures) != 2 {
		t.Fatalf("Expected 2 signatures (namespaced + bare), got %d", len(signatures))
	}

	if signatures[0].FunctionName != "ta.sma" {
		t.Errorf("First signature name = %q, want %q", signatures[0].FunctionName, "ta.sma")
	}
	if signatures[1].FunctionName != "sma" {
		t.Errorf("Second signature name = %q, want %q", signatures[1].FunctionName, "sma")
	}

	if signatures[0].DefaultSource != "close" {
		t.Errorf("Namespaced signature default source = %q, want %q", signatures[0].DefaultSource, "close")
	}
	if signatures[1].DefaultSource != "close" {
		t.Errorf("Bare signature default source = %q, want %q", signatures[1].DefaultSource, "close")
	}

	if len(signatures[0].Overloads) != 1 || len(signatures[1].Overloads) != 1 {
		t.Error("Both signatures should have identical overload rules")
	}
}

func TestAppendWithBareAlias_MultiNamespaceFunction(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)}),
	}

	signatures = appendWithBareAlias(signatures, "math.stats.avg", "close", overloads)

	if len(signatures) != 2 {
		t.Fatalf("Expected 2 signatures, got %d", len(signatures))
	}

	if signatures[0].FunctionName != "math.stats.avg" {
		t.Errorf("Namespaced name = %q, want %q", signatures[0].FunctionName, "math.stats.avg")
	}
	if signatures[1].FunctionName != "avg" {
		t.Errorf("Bare name = %q, want %q", signatures[1].FunctionName, "avg")
	}
}

func TestAppendWithBareAlias_NoNamespaceFunction(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewSeriesArgument(0, "")}),
	}

	signatures = appendWithBareAlias(signatures, "nz", "close", overloads)

	if len(signatures) != 1 {
		t.Fatalf("Bare function without dot should register once, got %d registrations", len(signatures))
	}

	if signatures[0].FunctionName != "nz" {
		t.Errorf("Function name = %q, want %q", signatures[0].FunctionName, "nz")
	}
}

func TestAppendWithBareAlias_EmptyDefaultSource(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarIntArgument(2),
		}),
	}

	signatures = appendWithBareAlias(signatures, "ta.pivothigh", "", overloads)

	if len(signatures) != 2 {
		t.Fatalf("Expected 2 signatures, got %d", len(signatures))
	}

	if signatures[0].DefaultSource != "" || signatures[1].DefaultSource != "" {
		t.Error("Both signatures should have empty default source")
	}
}

func TestAppendWithBareAlias_MultipleOverloads(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
		NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)}),
		NewSingleOverloadRule(3, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1), NewScalarIntArgument(2)}),
	}

	signatures = appendWithBareAlias(signatures, "ta.change", "close", overloads)

	if len(signatures) != 2 {
		t.Fatalf("Expected 2 signatures, got %d", len(signatures))
	}

	if len(signatures[0].Overloads) != 3 || len(signatures[1].Overloads) != 3 {
		t.Error("Both signatures should have all 3 overload rules")
	}

	for i := 0; i < 3; i++ {
		if signatures[0].Overloads[i].ArgCount != signatures[1].Overloads[i].ArgCount {
			t.Errorf("Overload %d arg count mismatch between namespaced and bare", i)
		}
	}
}

func TestAppendWithBareAlias_ChainedCalls(t *testing.T) {
	var signatures []TAFunctionMetadata
	overload1 := []TAOverloadRule{NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)})}
	overload2 := []TAOverloadRule{NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)})}

	signatures = appendWithBareAlias(signatures, "ta.sma", "close", overload1)
	signatures = appendWithBareAlias(signatures, "ta.ema", "close", overload2)

	if len(signatures) != 4 {
		t.Fatalf("Expected 4 signatures (2 functions × 2 forms), got %d", len(signatures))
	}

	names := make([]string, len(signatures))
	for i, sig := range signatures {
		names[i] = sig.FunctionName
	}

	expectedNames := []string{"ta.sma", "sma", "ta.ema", "ema"}
	for i, expected := range expectedNames {
		if names[i] != expected {
			t.Errorf("Signature[%d] name = %q, want %q", i, names[i], expected)
		}
	}
}

func TestAppendWithBareAlias_PreservesExistingSignatures(t *testing.T) {
	existingSignature := NewTAFunctionMetadata("existing.func", "high", []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
	})
	signatures := []TAFunctionMetadata{existingSignature}

	newOverloads := []TAOverloadRule{NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)})}
	signatures = appendWithBareAlias(signatures, "ta.new", "low", newOverloads)

	if len(signatures) != 3 {
		t.Fatalf("Expected 3 signatures (1 existing + 2 new), got %d", len(signatures))
	}

	if signatures[0].FunctionName != "existing.func" {
		t.Error("Existing signature should remain at index 0")
	}
	if signatures[0].DefaultSource != "high" {
		t.Error("Existing signature default source should be unchanged")
	}
}

func TestTAFunctionMetadata_FindOverload(t *testing.T) {
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
		NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)}),
	}
	metadata := NewTAFunctionMetadata("ta.sma", "close", overloads)

	tests := []struct {
		argCount  int
		wantFound bool
	}{
		{argCount: 1, wantFound: true},
		{argCount: 2, wantFound: true},
		{argCount: 0, wantFound: false},
		{argCount: 3, wantFound: false},
		{argCount: -1, wantFound: false},
	}

	for _, tt := range tests {
		t.Run(strings.Join([]string{"argCount", string(rune(tt.argCount + '0'))}, "_"), func(t *testing.T) {
			overload, found := metadata.FindOverload(tt.argCount)
			if found != tt.wantFound {
				t.Errorf("FindOverload(%d) found=%v, want %v", tt.argCount, found, tt.wantFound)
			}
			if found && overload.ArgCount != tt.argCount {
				t.Errorf("FindOverload(%d) returned overload with argCount=%d", tt.argCount, overload.ArgCount)
			}
		})
	}
}

func TestTAFunctionMetadata_SupportsArgCount(t *testing.T) {
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
		NewSingleOverloadRule(3, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1), NewScalarIntArgument(2)}),
	}
	metadata := NewTAFunctionMetadata("ta.pivothigh", "high", overloads)

	tests := []struct {
		argCount int
		want     bool
	}{
		{argCount: 1, want: true},
		{argCount: 2, want: false},
		{argCount: 3, want: true},
		{argCount: 0, want: false},
		{argCount: 4, want: false},
	}

	for _, tt := range tests {
		t.Run(strings.Join([]string{"argCount", string(rune(tt.argCount + '0'))}, "_"), func(t *testing.T) {
			got := metadata.SupportsArgCount(tt.argCount)
			if got != tt.want {
				t.Errorf("SupportsArgCount(%d) = %v, want %v", tt.argCount, got, tt.want)
			}
		})
	}
}

func TestTAFunctionMetadata_MinMaxArgCount(t *testing.T) {
	tests := []struct {
		name      string
		overloads []TAOverloadRule
		wantMin   int
		wantMax   int
	}{
		{
			name: "single_overload",
			overloads: []TAOverloadRule{
				NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)}),
			},
			wantMin: 2,
			wantMax: 2,
		},
		{
			name: "multiple_overloads",
			overloads: []TAOverloadRule{
				NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
				NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)}),
				NewSingleOverloadRule(3, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1), NewScalarIntArgument(2)}),
			},
			wantMin: 1,
			wantMax: 3,
		},
		{
			name: "non_sequential_overloads",
			overloads: []TAOverloadRule{
				NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
				NewSingleOverloadRule(3, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1), NewScalarIntArgument(2)}),
				NewSingleOverloadRule(5, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1), NewScalarIntArgument(2), NewScalarIntArgument(3), NewScalarIntArgument(4)}),
			},
			wantMin: 1,
			wantMax: 5,
		},
		{
			name:      "empty_overloads",
			overloads: []TAOverloadRule{},
			wantMin:   0,
			wantMax:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := NewTAFunctionMetadata("test.func", "close", tt.overloads)

			gotMin := metadata.MinArgCount()
			if gotMin != tt.wantMin {
				t.Errorf("MinArgCount() = %d, want %d", gotMin, tt.wantMin)
			}

			gotMax := metadata.MaxArgCount()
			if gotMax != tt.wantMax {
				t.Errorf("MaxArgCount() = %d, want %d", gotMax, tt.wantMax)
			}
		})
	}
}

func TestTAFunctionMetadata_OverloadIsolation(t *testing.T) {
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
		NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)}),
	}

	metadata1 := NewTAFunctionMetadata("ta.sma", "close", overloads)
	metadata2 := NewTAFunctionMetadata("ta.ema", "close", overloads)

	overload1, _ := metadata1.FindOverload(1)
	overload1.ArgCount = 99

	overload2, _ := metadata2.FindOverload(1)
	if overload2.ArgCount == 99 {
		t.Error("Modifying one metadata's overload should not affect another")
	}
}

func TestAppendTupleWithBareAlias_IsTuplePropagation(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(4, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarIntArgument(2),
			NewScalarIntArgument(3),
		}),
	}

	signatures = appendTupleWithBareAlias(signatures, "ta.macd", "close", overloads)

	if len(signatures) != 2 {
		t.Fatalf("Expected 2 signatures (namespaced + bare), got %d", len(signatures))
	}

	if !signatures[0].IsTuple {
		t.Error("Namespaced form must have IsTuple=true")
	}
	if !signatures[1].IsTuple {
		t.Error("Bare form must have IsTuple=true")
	}
}

func TestAppendTupleWithBareAlias_VsNonTuple(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(2, []TAArgumentSpec{NewSeriesArgument(0, ""), NewScalarIntArgument(1)}),
	}

	signatures = appendWithBareAlias(signatures, "ta.sma", "close", overloads)
	signatures = appendTupleWithBareAlias(signatures, "ta.macd", "close", overloads)

	for _, sig := range signatures {
		if sig.FunctionName == "ta.sma" || sig.FunctionName == "sma" {
			if sig.IsTuple {
				t.Errorf("%s registered via appendWithBareAlias must have IsTuple=false", sig.FunctionName)
			}
		}
		if sig.FunctionName == "ta.macd" || sig.FunctionName == "macd" {
			if !sig.IsTuple {
				t.Errorf("%s registered via appendTupleWithBareAlias must have IsTuple=true", sig.FunctionName)
			}
		}
	}
}

func TestAppendTupleWithBareAlias_NoNamespace(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(1, []TAArgumentSpec{NewScalarIntArgument(0)}),
	}

	signatures = appendTupleWithBareAlias(signatures, "custom_tuple", "", overloads)

	if len(signatures) != 1 {
		t.Fatalf("Function without dot should register once, got %d", len(signatures))
	}
	if !signatures[0].IsTuple {
		t.Error("Bare-only tuple function must have IsTuple=true")
	}
}

func TestAppendTupleWithBareAlias_PreservesMetadata(t *testing.T) {
	var signatures []TAFunctionMetadata
	overloads := []TAOverloadRule{
		NewSingleOverloadRule(3, []TAArgumentSpec{
			NewSeriesArgument(0, ""),
			NewScalarIntArgument(1),
			NewScalarFloatArgument(2),
		}),
	}

	signatures = appendTupleWithBareAlias(signatures, "ta.bb", "close", overloads)

	for _, sig := range signatures {
		if sig.DefaultSource != "close" {
			t.Errorf("%s DefaultSource = %q, want %q", sig.FunctionName, sig.DefaultSource, "close")
		}
		if len(sig.Overloads) != 1 {
			t.Errorf("%s should have 1 overload, got %d", sig.FunctionName, len(sig.Overloads))
		}
		if sig.Overloads[0].ArgCount != 3 {
			t.Errorf("%s overload ArgCount = %d, want 3", sig.FunctionName, sig.Overloads[0].ArgCount)
		}
	}
}

func TestNewTAFunctionMetadata_DefaultIsTupleFalse(t *testing.T) {
	meta := NewTAFunctionMetadata("ta.sma", "close", nil)
	if meta.IsTuple {
		t.Error("NewTAFunctionMetadata should default IsTuple to false")
	}
}
