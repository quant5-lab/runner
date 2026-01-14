package preprocessor

import "testing"

func TestNormalizeFunctionBlocks_NoArrowFunctions(t *testing.T) {
	input := `study(title="Test", shorttitle="Test", overlay=true)
ma20 = sma(close, 20)
plot(ma20, color=yellow, style=linebr, title="SMA20")`

	expected := input // Should remain unchanged
	result := NormalizeFunctionBlocks(input)

	if result != expected {
		t.Errorf("Non-arrow-function code should remain unchanged\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeFunctionBlocks_FunctionCallNotArrowFunc(t *testing.T) {
	input := `plot(value, color=red, title="Test")`
	expected := input
	result := NormalizeFunctionBlocks(input)

	if result != expected {
		t.Errorf("Function call should not be treated as arrow function\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}
