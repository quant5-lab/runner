package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type panicScenario struct {
	name            string
	script          string
	wantMsgFragment string
}

// runtimePanicScenarios exercises independent generator code paths so that
// recovery is proven path-agnostic, not tied to a single panic site.
var runtimePanicScenarios = []panicScenario{
	{
		name: "zero-step for-loop in statement",
		script: `//@version=5
indicator("Zero Step For Loop")
sum = 0.0
for i = 0 to 10 by 0
    sum := sum + i
plot(sum, "sum")
`,
		wantMsgFragment: "for loop step cannot be zero",
	},
	{
		name: "zero-step for-loop in arrow function body",
		script: `//@version=5
indicator("Arrow Zero Step")
f(src) =>
    total = 0.0
    for i = 0 to 5 by 0
        total := total + i
    total
plot(f(close), "f")
`,
		wantMsgFragment: "for loop step cannot be zero",
	},
}

func TestRuntimePanicRecovery_StructuredOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	for _, sc := range runtimePanicScenarios {
		t.Run(sc.name, func(t *testing.T) {
			testDir := t.TempDir()

			strategyPath := filepath.Join(testDir, "strategy.pine")
			if err := os.WriteFile(strategyPath, []byte(sc.script), 0644); err != nil {
				t.Fatal(err)
			}
			dataPath := filepath.Join(testDir, "data.json")
			if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(5, 86400)), 0644); err != nil {
				t.Fatal(err)
			}

			exePath := buildStrategyBinary(t, strategyPath, testDir, projectRoot)
			assertPanicRecovered(t, exePath, dataPath, testDir, sc.wantMsgFragment)
		})
	}
}

// TestRuntimePanicRecovery_NormalExecutionUnaffected is the critical negative case:
// the recovery mechanism must be transparent to strategies that complete without panicking.
func TestRuntimePanicRecovery_NormalExecutionUnaffected(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	script := `//@version=5
indicator("Normal Execution")
x = ta.sma(close, 3)
plot(x, "sma")
`
	testDir := t.TempDir()
	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))

	strategyPath := filepath.Join(testDir, "strategy.pine")
	if err := os.WriteFile(strategyPath, []byte(script), 0644); err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(testDir, "data.json")
	if err := os.WriteFile(dataPath, []byte(generateTestOHLCV(10, 86400)), 0644); err != nil {
		t.Fatal(err)
	}

	exePath := buildStrategyBinary(t, strategyPath, testDir, projectRoot)

	resultPath := filepath.Join(testDir, "result.json")
	runCmd := exec.Command(exePath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
	output, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("normal strategy exited non-zero (recover() must not intercept clean exits)\nexit: %v\noutput: %s", err, output)
	}
	if runCmd.ProcessState.ExitCode() != 0 {
		t.Errorf("exit code = %d, want 0 for normal execution", runCmd.ProcessState.ExitCode())
	}
}

func assertPanicRecovered(t *testing.T, exePath, dataPath, testDir, wantMsgFragment string) {
	t.Helper()

	resultPath := filepath.Join(testDir, "result.json")
	runCmd := exec.Command(exePath, "-symbol", "TEST", "-data", dataPath, "-output", resultPath)
	output, err := runCmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit from panic-inducing strategy, got success")
	}

	exitCode := runCmd.ProcessState.ExitCode()
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2", exitCode)
	}

	stderr := string(output)
	if !strings.Contains(stderr, "Runtime error:") {
		t.Errorf("stderr missing structured prefix %q\ngot: %q", "Runtime error:", stderr)
	}
	if !strings.Contains(stderr, wantMsgFragment) {
		t.Errorf("stderr missing panic message fragment %q\ngot: %q", wantMsgFragment, stderr)
	}
	if strings.Contains(stderr, "goroutine ") {
		t.Errorf("stderr contains raw Go goroutine dump — panic was not recovered\ngot: %q", stderr)
	}
}

func buildStrategyBinary(t *testing.T, strategyPath, testDir, projectRoot string) string {
	t.Helper()

	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	outputPlaceholder := filepath.Join(testDir, "strategy.go")
	genCmd := exec.Command(
		"go", "run", builderPath,
		"-input", strategyPath,
		"-output", outputPlaceholder,
		"-template", templatePath,
	)
	genOutput, err := genCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pine-gen failed: %v\n%s", err, genOutput)
	}

	generatedFile := resolveGeneratedFilePath(genOutput, outputPlaceholder)
	return buildFromGeneratedFile(t, generatedFile, testDir, projectRoot)
}

func resolveGeneratedFilePath(genOutput []byte, fallback string) string {
	const prefix = "Generated: "
	for _, line := range strings.Split(string(genOutput), "\n") {
		if strings.HasPrefix(line, prefix) {
			return filepath.Clean(strings.TrimPrefix(line, prefix))
		}
	}
	return fallback
}
