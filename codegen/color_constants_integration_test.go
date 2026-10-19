package codegen

import (
	"strings"
	"testing"
)

func TestColorConstants_BareIdentifiers(t *testing.T) {
	tests := []struct {
		name         string
		pine         string
		expectedHex  string
		variableName string
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
			expectedHex:  `"#4CAF50"`,
			variableName: "c1",
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
				t.Errorf("Expected %q in generated code", varDecl)
			}

			assignment := tt.variableName + " = " + tt.expectedHex
			if !strings.Contains(code, assignment) {
				t.Errorf("Expected %q in generated code", assignment)
			}
		})
	}
}

func TestColorConstants_MemberExpressions(t *testing.T) {
	tests := []struct {
		name         string
		pine         string
		expectedHex  string
		variableName string
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
				t.Errorf("Expected %q in generated code", varDecl)
			}

			assignment := tt.variableName + " = " + tt.expectedHex
			if !strings.Contains(code, assignment) {
				t.Errorf("Expected %q in generated code", assignment)
			}
		})
	}
}

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

func TestColorConstants_InPlotFunction(t *testing.T) {
	tests := []struct {
		name        string
		pine        string
		expectedHex string
	}{
		{
			name: "bare identifier in plot",
			pine: `
//@version=5
indicator("Test")
plot(close, color=blue)
`,
			expectedHex: `"#2962FF"`,
		},
		{
			name: "member expression in plot",
			pine: `
//@version=5
indicator("Test")
plot(close, color=color.red)
`,
			expectedHex: `"#FF5252"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			if !strings.Contains(code, tt.expectedHex) {
				t.Errorf("Expected %s in generated code", tt.expectedHex)
			}
		})
	}
}

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

func TestColorFunctions_ColorNew(t *testing.T) {
	tests := []struct {
		name     string
		pine     string
		expected []string
	}{
		{
			name: "color.new with member expression color",
			pine: `
//@version=5
indicator("Test")
myColor = color.new(color.red, 50)
`,
			expected: []string{
				"var myColor string",
				"visual.PineColorNew",
				"#FF5252",
			},
		},
		{
			name: "color.new with bare color identifier",
			pine: `
//@version=5
indicator("Test")
c = color.new(blue, 30)
`,
			expected: []string{
				"var c string",
				"visual.PineColorNew",
				"#2962FF",
			},
		},
		{
			name: "color.new with zero transparency",
			pine: `
//@version=5
indicator("Test")
solid = color.new(color.green, 0)
`,
			expected: []string{
				"var solid string",
				"visual.PineColorNew",
				"#4CAF50",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.pine)
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			for _, pattern := range tt.expected {
				if !strings.Contains(code, pattern) {
					t.Errorf("Missing pattern: %q\nGenerated code:\n%s", pattern, code)
				}
			}
		})
	}
}

func TestColorFunctions_ColorRGB(t *testing.T) {
	pine := `
//@version=5
indicator("Test")
custom = color.rgb(255, 128, 0, 20)
`

	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	expected := []string{
		"var custom string",
		"visual.PineColorRGB",
	}
	for _, pattern := range expected {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing pattern: %q\nGenerated code:\n%s", pattern, code)
		}
	}
}

func TestColorFunctions_InPlot(t *testing.T) {
	pine := `
//@version=5
indicator("Test")
plot(close, color=color.new(color.blue, 50))
`

	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	expected := []string{
		"visual.PineColorNew",
		"#2962FF",
	}
	for _, pattern := range expected {
		if !strings.Contains(code, pattern) {
			t.Errorf("Missing pattern: %q\nGenerated code:\n%s", pattern, code)
		}
	}
}

func TestColorConstants_NamespaceGuard(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "color.* in simple conditional assignment",
			script: `//@version=5
indicator("Color Guard")
bclr = close > open ? color.lime : color.red
barcolor(bclr)
`,
		},
		{
			name: "color.* in nested conditional (else-if)",
			script: `//@version=5
indicator("Color Guard Nested")
bclr = close > close[1] ? color.lime : (close < close[1] ? color.red : color.gray)
barcolor(bclr)
`,
		},
		{
			name: "color.* assigned to variable then used in barcolor",
			script: `//@version=5
indicator("Color Var")
upColor = color.green
downColor = color.red
bclr = close >= open ? upColor : downColor
barcolor(bclr)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := compilePineScript(tt.script)
			if err != nil {
				t.Fatalf("compile failed: %v", err)
			}
			if strings.Contains(code, "colorSeries") {
				t.Errorf("generated code contains undefined 'colorSeries':\n%s", code)
			}
		})
	}
}
