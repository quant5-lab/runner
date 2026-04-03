package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestArrayDrawingConstructors_ParseCodegenCompile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name string
		pine string
	}{
		{
			name: "empty_label_array",
			pine: `//@version=5
indicator("Label Array", overlay=true)
labels = array.new_label()
plot(close)
`,
		},
		{
			name: "sized_line_array",
			pine: `//@version=5
indicator("Line Array", overlay=true)
lines = array.new_line(10)
plot(close)
`,
		},
		{
			name: "empty_box_array",
			pine: `//@version=5
indicator("Box Array", overlay=true)
boxes = array.new_box()
plot(close)
`,
		},
		{
			name: "sized_table_array",
			pine: `//@version=5
indicator("Table Array", overlay=true)
tables = array.new_table(5)
plot(close)
`,
		},
		{
			name: "empty_linefill_array",
			pine: `//@version=5
indicator("Linefill Array", overlay=true)
fills = array.new_linefill()
plot(close)
`,
		},
		{
			name: "mixed_drawing_arrays",
			pine: `//@version=5
indicator("Mixed Drawing Arrays", overlay=true)
labels = array.new_label(3)
lines = array.new_line(5)
boxes = array.new_box()
tables = array.new_table(2)
fills = array.new_linefill(1)
plot(close)
`,
		},
		{
			name: "drawing_and_numeric_arrays",
			pine: `//@version=5
indicator("Mixed Array Types", overlay=true)
prices = array.new_float(10, close)
labels = array.new_label(5)
volumes = array.new_int(10)
lines = array.new_line()
flags = array.new_bool(3)
boxes = array.new_box(2)
plot(close)
`,
		},
	}

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pineFile := filepath.Join(tmpDir, tt.name+".pine")

			if err := os.WriteFile(pineFile, []byte(tt.pine), 0644); err != nil {
				t.Fatalf("write pine: %v", err)
			}

			outputGoPath := filepath.Join(tmpDir, tt.name+".go")
			genCmd := exec.Command(
				"go", "run", builderPath,
				"-input", pineFile,
				"-output", outputGoPath,
				"-template", templatePath,
			)
			genOutput, err := genCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("pine-gen failed: %v\n%s", err, genOutput)
			}

			generatedFile := ""
			for _, line := range strings.Split(string(genOutput), "\n") {
				if strings.HasPrefix(line, "Generated: ") {
					generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated: "))
					break
				}
			}
			if generatedFile == "" {
				t.Fatalf("failed to parse generated file path from output: %s", genOutput)
			}

			localGoFile := filepath.Join(tmpDir, "main.go")
			generatedData, err := os.ReadFile(generatedFile)
			if err != nil {
				t.Fatalf("read generated file: %v", err)
			}
			if err := os.WriteFile(localGoFile, generatedData, 0644); err != nil {
				t.Fatalf("write local go file: %v", err)
			}

			if err := setupGoMod(localGoFile, projectRoot); err != nil {
				t.Fatalf("setup go.mod: %v", err)
			}

			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = tmpDir
			if tidyOutput, err := tidyCmd.CombinedOutput(); err != nil {
				t.Fatalf("go mod tidy failed: %v\n%s", err, tidyOutput)
			}

			binPath := filepath.Join(tmpDir, tt.name)
			buildCmd := exec.Command("go", "build", "-o", binPath, localGoFile)
			buildCmd.Dir = tmpDir
			buildOutput, err := buildCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("go build failed: %v\n%s", err, buildOutput)
			}
		})
	}
}

func TestArrayDrawingConstructors_ZeroSizeEdgeCase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pine := `//@version=5
indicator("Zero Size Arrays", overlay=true)
labels = array.new_label(0)
lines = array.new_line(0)
boxes = array.new_box(0)
plot(close)
`

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, "zero_size.pine")
	if err := os.WriteFile(pineFile, []byte(pine), 0644); err != nil {
		t.Fatalf("write pine: %v", err)
	}

	outputGoPath := filepath.Join(tmpDir, "zero_size.go")
	genCmd := exec.Command(
		"go", "run", builderPath,
		"-input", pineFile,
		"-output", outputGoPath,
		"-template", templatePath,
	)
	genOutput, err := genCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pine-gen failed: %v\n%s", err, genOutput)
	}

	generatedFile := ""
	for _, line := range strings.Split(string(genOutput), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated: "))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("failed to parse generated file path from output: %s", genOutput)
	}

	localGoFile := filepath.Join(tmpDir, "main.go")
	generatedData, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if err := os.WriteFile(localGoFile, generatedData, 0644); err != nil {
		t.Fatalf("write local go file: %v", err)
	}

	if err := setupGoMod(localGoFile, projectRoot); err != nil {
		t.Fatalf("setup go.mod: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if tidyOutput, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, tidyOutput)
	}

	binPath := filepath.Join(tmpDir, "zero_size")
	buildCmd := exec.Command("go", "build", "-o", binPath, localGoFile)
	buildCmd.Dir = tmpDir
	if buildOutput, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, buildOutput)
	}
}

func TestArrayDrawingConstructors_DynamicSizeExpression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pine := `//@version=5
indicator("Dynamic Size", overlay=true)
sizeParam = 10
labels = array.new_label(sizeParam)
lines = array.new_line(sizeParam * 2)
boxes = array.new_box(sizeParam + 5)
plot(close)
`

	cwd, _ := os.Getwd()
	projectRoot := filepath.Dir(filepath.Dir(cwd))
	builderPath := filepath.Join(projectRoot, "cmd", "pine-gen", "main.go")
	templatePath := filepath.Join(projectRoot, "template", "main.go.tmpl")

	tmpDir := t.TempDir()
	pineFile := filepath.Join(tmpDir, "dynamic.pine")
	if err := os.WriteFile(pineFile, []byte(pine), 0644); err != nil {
		t.Fatalf("write pine: %v", err)
	}

	outputGoPath := filepath.Join(tmpDir, "dynamic.go")
	genCmd := exec.Command(
		"go", "run", builderPath,
		"-input", pineFile,
		"-output", outputGoPath,
		"-template", templatePath,
	)
	genOutput, err := genCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pine-gen failed: %v\n%s", err, genOutput)
	}

	generatedFile := ""
	for _, line := range strings.Split(string(genOutput), "\n") {
		if strings.HasPrefix(line, "Generated: ") {
			generatedFile = strings.TrimSpace(strings.TrimPrefix(line, "Generated: "))
			break
		}
	}
	if generatedFile == "" {
		t.Fatalf("failed to parse generated file path from output: %s", genOutput)
	}

	localGoFile := filepath.Join(tmpDir, "main.go")
	generatedData, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("read generated file: %v", err)
	}
	if err := os.WriteFile(localGoFile, generatedData, 0644); err != nil {
		t.Fatalf("write local go file: %v", err)
	}

	if err := setupGoMod(localGoFile, projectRoot); err != nil {
		t.Fatalf("setup go.mod: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if tidyOutput, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, tidyOutput)
	}

	binPath := filepath.Join(tmpDir, "dynamic")
	buildCmd := exec.Command("go", "build", "-o", binPath, localGoFile)
	buildCmd.Dir = tmpDir
	if buildOutput, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, buildOutput)
	}
}
