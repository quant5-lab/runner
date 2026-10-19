//go:build integration

package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// pipelineStage classifies where a strategy is expected to fail.
type pipelineStage int

const (
	stageCodegen pipelineStage = iota
	stageBuild
	stageAll
)

type strategyExpectation struct {
	skipStage pipelineStage
	reason    string
}

// skipList contains strategies that are known to fail at a specific pipeline stage.
// Each entry MUST have a reason explaining the root cause.
// When a strategy in this list starts passing all stages, the test fails to force
// removal of the stale entry.
var skipList = map[string]strategyExpectation{}

func TestStrategyPipelineSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("smoke test skipped in short mode")
	}

	_, thisFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	strategiesDir := filepath.Join(projectRoot, "strategies")

	pineFiles := collectPineFiles(t, strategiesDir)
	if len(pineFiles) == 0 {
		t.Fatal("no .pine files found under strategies/")
	}

	t.Logf("smoke-testing %d strategies", len(pineFiles))

	ohlcvFixture := []byte(generateTestOHLCV(50, 3600))

	for _, pineFile := range pineFiles {
		relPath := mustRelPath(t, strategiesDir, pineFile)
		t.Run(strings.ReplaceAll(relPath, string(filepath.Separator), "/"), func(t *testing.T) {
			t.Parallel()
			runStrategyPipeline(t, pineFile, relPath, projectRoot, ohlcvFixture)
		})
	}
}

func runStrategyPipeline(t *testing.T, pineFile, relPath, projectRoot string, ohlcvData []byte) {
	t.Helper()

	skip, isSkipped := skipList[relPath]

	workDir := t.TempDir()
	dataPath := filepath.Join(workDir, "data.json")
	if err := os.WriteFile(dataPath, ohlcvData, 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// Stage 1: codegen
	genFile, codegenOK := runCodegen(t, pineFile, workDir, projectRoot)
	if !codegenOK {
		if isSkipped && (skip.skipStage == stageCodegen || skip.skipStage == stageAll) {
			t.Skipf("known codegen failure (%s)", skip.reason)
		}
		t.Fatalf("codegen failed for non-skip-listed strategy")
	}
	if isSkipped && skip.skipStage == stageCodegen {
		t.Errorf("strategy is skip-listed for codegen failure but codegen succeeded — remove from skipList")
	}

	// Stage 2: build
	binaryPath, buildOK := runBuild(t, genFile, workDir, projectRoot)
	if !buildOK {
		if isSkipped && (skip.skipStage == stageBuild || skip.skipStage == stageAll) {
			t.Skipf("known build failure (%s)", skip.reason)
		}
		t.Fatalf("build failed for non-skip-listed strategy")
	}
	if isSkipped && skip.skipStage == stageBuild {
		t.Errorf("strategy is skip-listed for build failure but build succeeded — remove from skipList")
	}

	// Stage 3: execute
	resultPath := filepath.Join(workDir, "result.json")
	exitCode, stderr := runBinary(binaryPath, dataPath, resultPath)

	handleBinaryResult(t, exitCode, stderr)
}

func runCodegen(t *testing.T, pineFile, workDir, projectRoot string) (generatedPath string, ok bool) {
	t.Helper()

	outPath := filepath.Join(workDir, "strategy.go")
	pinegen := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	tmpl := filepath.Join(projectRoot, "template", "main.go.tmpl")

	cmd := exec.Command("go", "run", pinegen,
		"-input", pineFile,
		"-output", outPath,
		"-template", tmpl,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("codegen output:\n%s", out)
		return "", false
	}

	// Extract generated file path from output line "Generated: /path/to/file.go"
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			return strings.TrimPrefix(line, "Generated: "), true
		}
	}
	return outPath, true
}

func runBuild(t *testing.T, genFile, workDir, projectRoot string) (binaryPath string, ok bool) {
	t.Helper()

	localFile := filepath.Join(workDir, "strategy.go")
	if genFile != localFile {
		data, err := os.ReadFile(genFile)
		if err != nil {
			t.Logf("read generated file: %v", err)
			return "", false
		}
		if err := os.WriteFile(localFile, data, 0644); err != nil {
			t.Logf("copy generated file: %v", err)
			return "", false
		}
	}

	// Set up go.mod so the generated file can import runner packages
	if err := setupGoMod(localFile, projectRoot); err != nil {
		t.Logf("setupGoMod: %v", err)
		return "", false
	}
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = workDir
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		t.Logf("go mod tidy: %v\n%s", err, out)
		return "", false
	}

	exePath := filepath.Join(workDir, "strategy")
	buildCmd := exec.Command("go", "build", "-o", exePath, localFile)
	buildCmd.Dir = workDir
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Logf("build output:\n%s", out)
		return "", false
	}
	return exePath, true
}

func runBinary(binaryPath, dataPath, resultPath string) (exitCode int, stderr string) {
	cmd := exec.Command(binaryPath,
		"-symbol", "TEST",
		"-data", dataPath,
		"-output", resultPath,
	)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}
	return exitCode, stderrBuf.String()
}

func collectPineFiles(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".pine") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walkDir %s: %v", root, err)
	}
	return files
}

func mustRelPath(t *testing.T, base, target string) string {
	t.Helper()
	rel, err := filepath.Rel(base, target)
	if err != nil {
		t.Fatalf("relpath(%s, %s): %v", base, target, err)
	}
	return rel
}
