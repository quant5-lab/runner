package codegen

import (
	"strings"
	"testing"
)

/* TestColorConstants_BareIdentifiers validates bare color identifier compilation */
func TestColorConstants_BareIdentifiers(t *testing.T) {
	tests := []struct {
		name         string
		pine         string
		expectedHex  string
		variableName string
		description  string
	}{
		{
			name: "blue bare identifier",
			pine: `
//@version=5
indicator("Test")
myColor = blue
`,
			expectedHex:  `"#2962FF"`,
			variableName: "myColor",
			description:  "bare blue identifier resolves to TradingView blue hex",
		},
		{
			name: "red bare identifier",
			pine: `
//@version=5
indicator("Test")
alertColor = red
`,
			expectedHex:  `"#FF5252"`,
			variableName: "alertColor",
			description:  "bare red identifier resolves to TradingView red hex",
		},
		{
			name: "silver bare identifier",
			pine: `
//@version=5
indicator("Test")
neutralColor = silver
`,
			expectedHex:  `"#B2B5BE"`,
			variableName: "neutralColor",
			description:  "bare silver identifier resolves to TradingView silver hex",
		},
		{
			name: "multiple bare identifiers",
			pine: `
//@version=5
indicator("Test")
c1 = green
c2 = lime
c3 = maroon
`,
			expectedHex:  `"#4CAF50"`, // green
			variableName: "c1",
			description:  "multiple bare identifiers all resolve correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			varDecl := "var " + tt.variableName + " string"
			if !strings.Contains(code, varDecl) {
				t.Errorf("%s: Expected string variable declaration\n  Looking for: %s", tt.description, varDecl)
			}

			assignment := tt.variableName + " = " + tt.expectedHex
			if !strings.Contains(code, assignment) {
				t.Errorf("%s: Expected hex assignment\n  Looking for: %s", tt.description, assignment)
			}
		})
	}
}

/* TestColorConstants_MemberExpressions validates color.* member expression compilation */
func TestColorConstants_MemberExpressions(t *testing.T) {
	tests := []struct {
		name         string
		pine         string
		expectedHex  string
		variableName string
		description  string
	}{
		{
			name: "color.blue member expression",
			pine: `
//@version=5
indicator("Test")
plotColor = color.blue
`,
			expectedHex:  `"#2962FF"`,
			variableName: "plotColor",
			description:  "color.blue member expression resolves to hex",
		},
		{
			name: "color.red member expression",
			pine: `
//@version=5
indicator("Test")
trendColor = color.red
`,
			expectedHex:  `"#FF5252"`,
			variableName: "trendColor",
			description:  "color.red member expression resolves to hex",
		},
		{
			name: "color.aqua member expression",
			pine: `
//@version=5
indicator("Test")
bgColor = color.aqua
`,
			expectedHex:  `"#00BCD4"`,
			variableName: "bgColor",
			description:  "color.aqua member expression resolves to hex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			varDecl := "var " + tt.variableName + " string"
			if !strings.Contains(code, varDecl) {
				t.Errorf("%s: Expected string variable declaration\n  Looking for: %s", tt.description, varDecl)
			}

			assignment := tt.variableName + " = " + tt.expectedHex
			if !strings.Contains(code, assignment) {
				t.Errorf("%s: Expected hex assignment\n  Looking for: %s", tt.description, assignment)
			}
		})
	}
}

/* TestColorConstants_AllSupported validates all 17 TradingView colors compile */
func TestColorConstants_AllSupported(t *testing.T) {
	allColors := []struct {
		name string
		hex  string
	}{
		{"aqua", `"#00BCD4"`},
		{"black", `"#363A45"`},
		{"blue", `"#2962FF"`},
		{"fuchsia", `"#E040FB"`},
		{"gray", `"#787B86"`},
		{"green", `"#4CAF50"`},
		{"lime", `"#00E676"`},
		{"maroon", `"#880E4F"`},
		{"navy", `"#311B92"`},
		{"olive", `"#808000"`},
		{"orange", `"#FF9800"`},
		{"purple", `"#9C27B0"`},
		{"red", `"#FF5252"`},
		{"silver", `"#B2B5BE"`},
		{"teal", `"#00897B"`},
		{"white", `"#FFFFFF"`},
		{"yellow", `"#FFEB3B"`},
	}

	t.Run("bare identifiers", func(t *testing.T) {
		var pineScript strings.Builder
		pineScript.WriteString("//@version=5\nindicator(\"All Colors\")\n")
		for _, color := range allColors {
			pineScript.WriteString("c_" + color.name + " = " + color.name + "\n")
		}

		code, err := compilePineScript(pineScript.String())
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}

		for _, color := range allColors {
			varName := "c_" + color.name
			varDecl := "var " + varName + " string"
			if !strings.Contains(code, varDecl) {
				t.Errorf("Missing string declaration for %s: expected %s", color.name, varDecl)
			}

			assignment := varName + " = " + color.hex
			if !strings.Contains(code, assignment) {
				t.Errorf("Missing bare identifier assignment for %s: expected %s", color.name, assignment)
			}
		}
	})

	t.Run("member expressions", func(t *testing.T) {
		var pineScript strings.Builder
		pineScript.WriteString("//@version=5\nindicator(\"All Colors Member\")\n")
		for _, color := range allColors {
			pineScript.WriteString("m_" + color.name + " = color." + color.name + "\n")
		}

		code, err := compilePineScript(pineScript.String())
		if err != nil {
			t.Fatalf("Compilation failed: %v", err)
		}

		for _, color := range allColors {
			varName := "m_" + color.name
			varDecl := "var " + varName + " string"
			if !strings.Contains(code, varDecl) {
				t.Errorf("Missing string declaration for %s: expected %s", color.name, varDecl)
			}

			assignment := varName + " = " + color.hex
			if !strings.Contains(code, assignment) {
				t.Errorf("Missing member expression assignment for %s: expected %s", color.name, assignment)
			}
		}
	})
}

/* TestColorConstants_InPlotFunction validates color constants in plot() calls */
func TestColorConstants_InPlotFunction(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		expectedHex string
		description string
	}{
		{
			name: "bare identifier in plot",
			pine: `
//@version=5
indicator("Test")
plot(close, color=blue)
`,
			expectedHex: `"#2962FF"`,
			description: "bare blue in plot color parameter",
		},
		{
			name: "member expression in plot",
			pine: `
//@version=5
indicator("Test")
plot(close, color=color.red)
`,
			expectedHex: `"#FF5252"`,
			description: "color.red in plot color parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			if !strings.Contains(code, tt.expectedHex) {
				t.Errorf("%s: Expected hex value %s in generated code", tt.description, tt.expectedHex)
			}
		})
	}
}

/* TestColorConstants_MixedUsage validates bare and member expressions in same script */
func TestColorConstants_MixedUsage(t *testing.T) {
	pine := `
//@version=5
indicator("Mixed Color Usage")

bareBlue = blue
memberRed = color.red
conditionalColor = close > open ? green : color.silver

plot(close, color=bareBlue)
plot(high, color=memberRed)
plot(low, color=conditionalColor)
`

	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	requiredPatterns := []struct {
		pattern     string
		description string
	}{
		{`var bareBlue string`, "bareBlue string declaration"},
		{`bareBlue = "#2962FF"`, "bare blue variable assignment"},
		{`var memberRed string`, "memberRed string declaration"},
		{`memberRed = "#FF5252"`, "member red variable assignment"},
		{`"#4CAF50"`, "green in conditional (consequent)"},
		{`"#B2B5BE"`, "silver in conditional (alternate)"},
	}

	for _, req := range requiredPatterns {
		if !strings.Contains(code, req.pattern) {
			t.Errorf("Missing required pattern: %s\n  Pattern: %s", req.description, req.pattern)
		}
	}
}
