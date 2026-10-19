package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// codegenBuildResult carries the outputs of a successful codegen+build pipeline run.
type codegenBuildResult struct {
	BinaryPath    string
	GeneratedPath string
}

// codegenAndBuild runs pine-gen on the given Pine source and compiles the result.
// Returns (result, false) if codegen or compilation fails; t.Fatal is called for
// infrastructure errors (file I/O, go mod tidy) that are never expected to fail.
func codegenAndBuild(t *testing.T, tmpDir, name, pineSource, projectRoot string) (result codegenBuildResult, ok bool) {
	t.Helper()

	pineFile := filepath.Join(tmpDir, name+".pine")
	if err := os.WriteFile(pineFile, []byte(pineSource), 0644); err != nil {
		t.Fatalf("write pine: %v", err)
	}

	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")
	outputGoPath := filepath.Join(tmpDir, name+".go")

	genCmd := exec.Command("go", "run", builderPath,
		"-input", pineFile,
		"-output", outputGoPath,
		"-template", templatePath,
	)
	genOutput, err := genCmd.CombinedOutput()
	if err != nil {
		t.Logf("pine-gen output:\n%s", genOutput)
		return result, false
	}

	generatedFile := ""
	for _, line := range strings.Split(string(genOutput), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated: "))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("cannot parse generated file path from pine-gen output: %s", genOutput)
	}

	localGoFile := filepath.Join(tmpDir, "main.go")
	generatedData, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if err := os.WriteFile(localGoFile, generatedData, 0644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	if err := setupGoMod(localGoFile, projectRoot); err != nil {
		t.Fatalf("setup go.mod: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	binPath := filepath.Join(tmpDir, name)
	buildCmd := exec.Command("go", "build", "-o", binPath, localGoFile)
	buildCmd.Dir = tmpDir
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("go build output:\n%s", buildOutput)
		return result, false
	}

	return codegenBuildResult{BinaryPath: binPath, GeneratedPath: generatedFile}, true
}

// projectRootFromCwd returns the runner module root for test helpers that need to
// invoke cmd/pine-gen or reference the template directory.
func projectRootFromCwd() string {
	cwd, _ := os.Getwd()
	return filepath.Dir(filepath.Dir(cwd))
}
