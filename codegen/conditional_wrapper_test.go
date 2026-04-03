package codegen

import (
	"strings"
	"testing"
)

func TestConditionalWrapperGenerator_NoCondition(t *testing.T) {
	wrapper := &ConditionalWrapperGenerator{}

	bodyCode := "\t\tstrat.Entry(\"Long\", strategy.Long, 1.0, \"\")\n"
	result := wrapper.WrapIfNeeded("", bodyCode, "\t\t")

	if result != bodyCode {
		t.Errorf("Expected original body when no condition, got:\n%s", result)
	}
}

func TestConditionalWrapperGenerator_SimpleCondition(t *testing.T) {
	wrapper := &ConditionalWrapperGenerator{}

	bodyCode := "\t\tstrat.Entry(\"Long\", strategy.Long, 1.0, \"\")\n"
	condition := "bar.Close > bar.Open"
	result := wrapper.WrapIfNeeded(condition, bodyCode, "\t\t")

	expected := "\t\tif value.IsTrue(bar.Close > bar.Open) {\n" +
		"\t\t\tstrat.Entry(\"Long\", strategy.Long, 1.0, \"\")\n" +
		"\t\t}\n"

	if result != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestConditionalWrapperGenerator_MultiLineBody(t *testing.T) {
	wrapper := &ConditionalWrapperGenerator{}

	bodyCode := "\t\tentryQty := 1000 / closeSeries.GetCurrent()\n" +
		"\t\tstrat.Entry(\"Long\", strategy.Long, entryQty, \"\")\n"
	condition := "smaFast > smaSlow"
	result := wrapper.WrapIfNeeded(condition, bodyCode, "\t\t")

	lines := strings.Split(result, "\n")
	if len(lines) < 4 {
		t.Fatalf("Expected at least 4 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "if value.IsTrue(smaFast > smaSlow) {") {
		t.Errorf("Expected if-block opening, got: %s", lines[0])
	}

	if !strings.Contains(lines[1], "entryQty := 1000") {
		t.Errorf("Expected first body line indented, got: %s", lines[1])
	}

	if !strings.Contains(lines[2], "strat.Entry") {
		t.Errorf("Expected second body line indented, got: %s", lines[2])
	}

	if !strings.Contains(lines[3], "}") {
		t.Errorf("Expected closing brace, got: %s", lines[3])
	}
}

func TestConditionalWrapperGenerator_ComplexCondition(t *testing.T) {
	wrapper := &ConditionalWrapperGenerator{}

	bodyCode := "\t\tstrat.Entry(\"Short\", strategy.Short, 2.0, \"Reversal\")\n"
	condition := "value.IsTrue(func() float64 { if (smaFast < smaSlow && rsiSeries.Get(0) > 70) { return 1.0 } else { return 0.0 } }())"
	result := wrapper.WrapIfNeeded(condition, bodyCode, "\t\t")

	if !strings.HasPrefix(result, "\t\tif value.IsTrue(") {
		t.Errorf("Expected if-block prefix, got: %s", result[:30])
	}

	if !strings.Contains(result, "strat.Entry") {
		t.Error("Expected Entry call in wrapped body")
	}

	if !strings.HasSuffix(strings.TrimSpace(result), "}") {
		t.Errorf("Expected closing brace at end, got: %s", result[len(result)-10:])
	}
}

func TestConditionalWrapperGenerator_PreservesIndentation(t *testing.T) {
	wrapper := &ConditionalWrapperGenerator{}

	tests := []struct {
		name   string
		indent string
		body   string
	}{
		{
			name:   "no indentation",
			indent: "",
			body:   "strat.Entry(\"Long\", strategy.Long, 1.0, \"\")\n",
		},
		{
			name:   "one tab",
			indent: "\t",
			body:   "\tstrat.Entry(\"Long\", strategy.Long, 1.0, \"\")\n",
		},
		{
			name:   "two tabs",
			indent: "\t\t",
			body:   "\t\tstrat.Entry(\"Long\", strategy.Long, 1.0, \"\")\n",
		},
		{
			name:   "three tabs",
			indent: "\t\t\t",
			body:   "\t\t\tstrat.Entry(\"Long\", strategy.Long, 1.0, \"\")\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapper.WrapIfNeeded("condition", tt.body, tt.indent)

			lines := strings.Split(result, "\n")
			if !strings.HasPrefix(lines[0], tt.indent+"if") {
				t.Errorf("Expected if-line to start with %q, got: %s", tt.indent, lines[0])
			}

			if !strings.HasPrefix(lines[len(lines)-2], tt.indent+"}") {
				t.Errorf("Expected closing brace to start with %q, got: %s", tt.indent, lines[len(lines)-2])
			}
		})
	}
}
