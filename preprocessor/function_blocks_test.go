package preprocessor

import (
	"strings"
	"testing"
)

func TestNormalizeFunctionBlocks_SingleFunction(t *testing.T) {
	input := `funcA(x) =>
    result = x + 1
    result`

	expected := `funcA(x) => @BEGIN
    result = x + 1
    result
@END`

	result := NormalizeFunctionBlocks(input)
	if result != expected {
		t.Errorf("Single function normalization failed\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeFunctionBlocks_MultipleFunctions(t *testing.T) {
	input := `//@version=4
study("Test", overlay=false)

funcA(x) =>
    a = x + 1
    b = x + 2
    [a, b]

funcB(val) =>
    [r1, r2] = funcA(val)
    r1 + r2

result = funcB(10)
plot(result)`

	result := NormalizeFunctionBlocks(input)

	lines := strings.Split(result, "\n")

	funcAFound := false
	funcBFound := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "funcA(x) =>") {
			funcAFound = true
			if i+3 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if nextLine != "a = x + 1" {
					t.Errorf("funcA body line 1 incorrect: got %s", nextLine)
				}
			}
		}
		if strings.HasPrefix(trimmed, "funcB(val) =>") {
			funcBFound = true
			if !funcAFound {
				t.Error("funcB should appear after funcA")
			}
		}
	}

	if !funcAFound {
		t.Error("funcA not found in output")
	}
	if !funcBFound {
		t.Error("funcB not found in output")
	}
}

func TestNormalizeFunctionBlocks_BB7Pattern(t *testing.T) {
	input := `dirmov(len) =>
    up = change(high)
    down = -change(low)
    truerange = rma(tr, len)
    plus = fixnan(100 * rma(up > down and up > 0 ? up : 0, len) / truerange)
    minus = fixnan(100 * rma(down > up and down > 0 ? down : 0, len) / truerange)
    [plus, minus]

adx(LWdilength, LWadxlength) =>
    [plus, minus] = dirmov(LWdilength)
    sum = plus + minus
    adx = 100 * rma(abs(plus - minus) / (sum == 0 ? 1 : sum), LWadxlength)
    [adx, plus, minus]`

	result := NormalizeFunctionBlocks(input)

	lines := strings.Split(result, "\n")

	dirmovFound := false
	adxFound := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "dirmov(len) =>") {
			dirmovFound = true
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if nextLine != "up = change(high)" {
					t.Errorf("dirmov first body line incorrect: got %s", nextLine)
				}
			}
		}
		if strings.HasPrefix(trimmed, "adx(LWdilength, LWadxlength) =>") {
			adxFound = true
			if !dirmovFound {
				t.Error("adx should appear after dirmov")
			}
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if nextLine != "[plus, minus] = dirmov(LWdilength)" {
					t.Errorf("adx first body line incorrect: got %s", nextLine)
				}
			}
		}
	}

	if !dirmovFound {
		t.Error("dirmov function not found in output")
	}
	if !adxFound {
		t.Error("adx function not found in output")
	}
}

func TestNormalizeFunctionBlocks_WithComments(t *testing.T) {
	input := `// Helper function
funcA(x) =>
    // Calculate result
    result = x + 1
    result

// Main function
funcB(val) =>
    funcA(val)`

	result := NormalizeFunctionBlocks(input)

	if !strings.Contains(result, "funcA(x) =>") {
		t.Error("funcA declaration missing")
	}
	if !strings.Contains(result, "funcB(val) =>") {
		t.Error("funcB declaration missing")
	}
	if !strings.Contains(result, "// Helper function") {
		t.Error("Comment before funcA missing")
	}
}

func TestNormalizeFunctionBlocks_NoFunctions(t *testing.T) {
	input := `//@version=4
study("Test", overlay=false)
sma20 = ta.sma(close, 20)
plot(sma20)`

	expected := input
	result := NormalizeFunctionBlocks(input)

	if result != expected {
		t.Errorf("Non-function code should remain unchanged\nExpected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestNormalizeFunctionBlocks_EmptyLinesInBody(t *testing.T) {
	input := `funcA(x) =>
    a = x + 1

    b = x + 2
    [a, b]`

	result := NormalizeFunctionBlocks(input)

	if !strings.Contains(result, "a = x + 1") {
		t.Error("First statement missing")
	}
	if !strings.Contains(result, "b = x + 2") {
		t.Error("Second statement missing")
	}
	if !strings.Contains(result, "[a, b]") {
		t.Error("Return statement missing")
	}
}

func TestNormalizeFunctionBlocks_MultipleParams(t *testing.T) {
	input := `funcMulti(a, b, c) =>
    result = a + b + c
    result`

	result := NormalizeFunctionBlocks(input)

	if !strings.Contains(result, "funcMulti(a, b, c) =>") {
		t.Error("Function with multiple params not preserved")
	}
	if !strings.Contains(result, "result = a + b + c") {
		t.Error("Function body missing")
	}
}

func TestNormalizeFunctionBlocks_IndentedFunction(t *testing.T) {
	input := `if condition
    helperFunc(x) =>
        x + 1`

	result := NormalizeFunctionBlocks(input)

	if !strings.Contains(result, "helperFunc(x) =>") {
		t.Error("Indented function declaration missing")
	}
}
