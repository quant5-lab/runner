package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type BlockerTestCase struct {
	name           string
	pineFile       string
	expectParse    bool
	expectGenerate bool
	expectCompile  bool
	errorContains  string
}

/* TestPineScriptBlockers validates documented blockers preventing 100% PineScript support */
func TestPineScriptBlockers(t *testing.T) {
	tests := []BlockerTestCase{
		{
			name:           "for_loops_parse_but_generate_literals",
			pineFile:       "test-for-loop.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  false,
			errorContains:  "not used",
		},
		{
			name:           "while_loops_parse_error",
			pineFile:       "test-while-loop.pine",
			expectParse:    false,
			expectGenerate: false,
			expectCompile:  false,
			errorContains:  "binary expression should be used in condition context",
		},
		{
			name:           "var_declarations_work",
			pineFile:       "test-var-decl.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "label_functions_work",
			pineFile:       "test-label.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "array_functions_work",
			pineFile:       "test-array.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "strategy_exit_works",
			pineFile:       "test-strategy-exit.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "bitwise_operators_lexer_error",
			pineFile:       "test-operators.pine",
			expectParse:    false,
			expectGenerate: false,
			expectCompile:  false,
			errorContains:  "lexer: invalid input text",
		},
		{
			name:           "ta_functions_cci_wma_vwap_work",
			pineFile:       "test-ta-missing.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "color_functions_work",
			pineFile:       "test-color-funcs.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "visual_functions_work",
			pineFile:       "test-visual-funcs.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "alert_functions_codegen_todo",
			pineFile:       "test-alert.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "string_functions_codegen_todo",
			pineFile:       "test-string-funcs.pine",
			expectParse:    true,
			expectGenerate: true,
			expectCompile:  true,
			errorContains:  "",
		},
		{
			name:           "map_generics_parse_error",
			pineFile:       "test-map.pine",
			expectParse:    false,
			expectGenerate: false,
			expectCompile:  false,
			errorContains:  "unexpected token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			originalDir, _ := os.Getwd()
			os.Chdir("../..")
			defer os.Chdir(originalDir)

			pineFilePath := filepath.Join("testdata/blockers", tt.pineFile)
			outputBinary := filepath.Join(tmpDir, "test_binary")

			buildCmd := exec.Command("go", "run", "cmd/pine-gen/main.go",
				"-input", pineFilePath,
				"-output", outputBinary)

			buildOutput, err := buildCmd.CombinedOutput()
			outputStr := string(buildOutput)

			if !tt.expectParse {
				if err == nil {
					t.Errorf("Expected parse error but succeeded")
				}
				if tt.errorContains != "" && !strings.Contains(outputStr, tt.errorContains) {
					t.Errorf("Expected error containing %q, got: %s", tt.errorContains, outputStr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse failed unexpectedly: %v\nOutput: %s", err, outputStr)
			}

			if !tt.expectGenerate {
				t.Fatalf("Generated code when generation should fail")
			}

			tempGoFile := ParseGeneratedFilePath(t, buildOutput)
			if tempGoFile == "" {
				t.Fatalf("Could not find generated Go file path")
			}

			generatedCode, err := os.ReadFile(tempGoFile)
			if err != nil {
				t.Fatalf("Failed to read generated code: %v", err)
			}

			binaryPath := filepath.Join(tmpDir, "test_binary")
			compileCmd := exec.Command("go", "build", "-o", binaryPath, tempGoFile)
			compileOutput, err := compileCmd.CombinedOutput()

			if !tt.expectCompile {
				if err == nil {
					t.Error("Expected compilation to fail but it succeeded")
				}
				if tt.errorContains != "" && !strings.Contains(string(compileOutput), tt.errorContains) {
					t.Errorf("Expected compile error containing %q, got: %s", tt.errorContains, compileOutput)
				}
				return
			}

			if err != nil {
				t.Fatalf("Compilation failed: %v\nOutput: %s", err, compileOutput)
			}

			verifyBlockerBehavior(t, tt.pineFile, string(generatedCode))
		})
	}
}

/* verifyBlockerBehavior checks implementation status of generated code */
func verifyBlockerBehavior(t *testing.T, pineFile string, generatedCode string) {
	switch pineFile {
	case "test-for-loop.pine":
		if !strings.Contains(generatedCode, "sumVal = 50.0") {
			t.Error("for loop should generate literal value, not loop execution")
		}
	case "test-alert.pine":
		if !strings.Contains(generatedCode, "// alert() - TODO: implement") {
			t.Error("alert() should have TODO comment")
		}
		if !strings.Contains(generatedCode, "// alertcondition() - TODO: implement") {
			t.Error("alertcondition() should have TODO comment")
		}
	case "test-string-funcs.pine":
		if !strings.Contains(generatedCode, "// str.") && !strings.Contains(generatedCode, "TODO") {
			t.Error("string functions should have TODO comments")
		}
	}
}

/* TestBlockerDocumentation ensures BLOCKERS.md stays synchronized with test results */
func TestBlockerDocumentation(t *testing.T) {
	blockersPath := "../../../docs/BLOCKERS.md"
	content, err := os.ReadFile(blockersPath)
	if err != nil {
		t.Fatalf("Failed to read BLOCKERS.md: %v", err)
	}

	blockersDoc := string(content)

	requiredSections := []string{
		"## CODEGEN LIMITATIONS",
		"## PARSER LIMITATIONS",
		"## BUILT-IN FUNCTIONS",
		"## OPERATORS",
		"## SUMMARY",
	}

	for _, section := range requiredSections {
		if !strings.Contains(blockersDoc, section) {
			t.Errorf("BLOCKERS.md missing required section: %s", section)
		}
	}

	if !strings.Contains(blockersDoc, "**Documented Blockers:**") {
		t.Error("BLOCKERS.md missing summary count")
	}

	expectedBlockers := []string{
		"RSI inline",
		"while loops",
		"map generics",
		"bitwise operators",
	}

	for _, blocker := range expectedBlockers {
		if !strings.Contains(blockersDoc, blocker) {
			t.Errorf("BLOCKERS.md missing documented blocker: %s", blocker)
		}
	}
}
