package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/parser"
)

func parseScript(t *testing.T, script string) string {
	t.Helper()

	p, err := parser.NewParser()
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	parseResult, err := p.ParseBytes("test.pine", []byte(script))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	converter := parser.NewConverter()
	program, err := converter.ToESTree(parseResult)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	code, err := GenerateStrategyCodeFromAST(program)
	if err != nil {
		t.Fatalf("GenerateStrategyCodeFromAST() error = %v", err)
	}

	return code.FunctionBody
}

func TestSecurityEvaluator_VarLookupConfiguration(t *testing.T) {
	tests := []struct {
		name              string
		script            string
		expectedVarInCase string
	}{
		{
			name: "inline security with binary expression referencing user variable",
			script: `//@version=4
strategy("Test", overlay=true)
threshold = sma(close, 20)
signal = security(syminfo.tickerid, "1D", close > threshold) ? 1 : 0
plot(signal)`,
			expectedVarInCase: "threshold",
		},
		{
			name: "top-level security with user variable in TA argument",
			script: `//@version=4
strategy("Test", overlay=true)
period = 14
rsi_val = security(syminfo.tickerid, "1D", rsi(close, period))
plot(rsi_val)`,
			expectedVarInCase: "period",
		},
		{
			name: "multiple user variables in security expression",
			script: `//@version=4
strategy("Test", overlay=true)
upper = sma(close, 20) + stdev(close, 20)
lower = sma(close, 20) - stdev(close, 20)
result = security(syminfo.tickerid, "1D", (close - upper) / (upper - lower))
plot(result)`,
			expectedVarInCase: "upper",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatedCode := parseScript(t, tt.script)

			if !strings.Contains(generatedCode, "SetVarLookup") {
				t.Error("Security evaluator must configure VarLookup for variable resolution")
			}

			expectedCase := `case "` + tt.expectedVarInCase + `":`
			if !strings.Contains(generatedCode, expectedCase) {
				t.Errorf("VarLookup switch must include case for user variable %q", tt.expectedVarInCase)
			}

			if !strings.Contains(generatedCode, tt.expectedVarInCase+"Series") {
				t.Errorf("VarLookup must access Series storage for variable %q", tt.expectedVarInCase)
			}
		})
	}
}

func TestSecurityEvaluator_BarIndexMapperConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "inline security requiring bar mapping",
			script: `//@version=4
strategy("Test", overlay=true)
myVar = ema(close, 10)
signal = security(syminfo.tickerid, "1D", close > myVar) ? 1 : 0
plot(signal)`,
		},
		{
			name: "nested security in conditional requiring bar mapping",
			script: `//@version=4
strategy("Test", overlay=true)
threshold = sma(close, 20)
val2 = close > threshold ? security(syminfo.tickerid, "1D", high > threshold) : 0
plot(val2)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatedCode := parseScript(t, tt.script)

			if !strings.Contains(generatedCode, "SetBarIndexMapper") {
				t.Error("Security evaluator must configure BarIndexMapper for temporal alignment")
			}

			if !strings.Contains(generatedCode, "NewBarIndexMapper") {
				t.Error("BarIndexMapper must be instantiated for security evaluator")
			}

			if !strings.Contains(generatedCode, "barMapper.SetMapping") {
				t.Error("BarIndexMapper must populate mappings from securityBarMapper ranges")
			}
		})
	}
}

func TestSecurityEvaluator_InputConstantsConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "security with TA function using input parameter",
			script: `//@version=4
strategy("Test", overlay=true)
length = input(14, "Length")
rsi_1d = security(syminfo.tickerid, "1D", rsi(close, length))
plot(rsi_1d)`,
		},
		{
			name: "inline security with input-driven comparison",
			script: `//@version=4
strategy("Test", overlay=true)
threshold = input(50.0, "Threshold")
signal = security(syminfo.tickerid, "1D", rsi(close, 14) > threshold) ? 1 : 0
plot(signal)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatedCode := parseScript(t, tt.script)

			if !strings.Contains(generatedCode, "SetInputConstantsMap") {
				t.Error("Security evaluator must configure InputConstantsMap for input() resolution")
			}

			if !strings.Contains(generatedCode, "inputConstantsMap") {
				t.Error("InputConstantsMap variable must be declared for evaluator configuration")
			}
		})
	}
}

func TestSecurityEvaluator_PersistentEvaluatorReuse(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "inline security reuses persistent evaluator",
			script: `//@version=4
strategy("Test", overlay=true)
myVar = sma(close, 10)
signal = security(syminfo.tickerid, "1D", close > myVar) ? 1 : 0
plot(signal)`,
		},
		{
			name: "multiple security contexts share evaluator instance",
			script: `//@version=4
strategy("Test", overlay=true)
threshold = sma(close, 20)
signal1 = close > 50 ? security(syminfo.tickerid, "1D", close > threshold) : 0
signal2 = close < 50 ? security(syminfo.tickerid, "1W", high > threshold) : 0
plot(signal1 + signal2)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatedCode := parseScript(t, tt.script)

			if !strings.Contains(generatedCode, "if secBarEvaluator == nil") {
				t.Error("Evaluator must check for existing instance before initialization")
			}

			if !strings.Contains(generatedCode, "secBarEvaluator.EvaluateAtBar") {
				t.Error("Evaluator must reference persistent secBarEvaluator instance")
			}

			newEvaluatorCount := strings.Count(generatedCode, "security.NewSeriesCachingEvaluator")
			if newEvaluatorCount == 0 {
				t.Error("Evaluator must be instantiated at least once")
			}
		})
	}
}

func TestSecurityEvaluator_ConsistentSetupAcrossContexts(t *testing.T) {
	script := `//@version=4
strategy("Test", overlay=true)

myVar = ema(close, 14)

result1 = security(syminfo.tickerid, "1D", sma(close, 20))
result2 = security(syminfo.tickerid, "1D", close > myVar) ? 1 : 0

getIndicator() =>
    security(syminfo.tickerid, "1W", rsi(close, 14))

result3 = getIndicator()
plot(result1 + result2 + result3)
`

	generatedCode := parseScript(t, script)

	requiredSetupCalls := []string{
		"SetVarLookup",
		"SetBarIndexMapper",
		"SetInputConstantsMap",
		"SetVariableRegistry",
	}

	for _, setupCall := range requiredSetupCalls {
		if !strings.Contains(generatedCode, setupCall) {
			t.Errorf("All security contexts must emit %s configuration", setupCall)
		}
	}

	if !strings.Contains(generatedCode, `case "myVar":`) {
		t.Error("VarLookup must include all user-defined variables from symbol table")
	}

	evaluatorInitCount := strings.Count(generatedCode, "if secBarEvaluator == nil")
	if evaluatorInitCount == 0 {
		t.Error("At least one security context must initialize evaluator")
	}
}
