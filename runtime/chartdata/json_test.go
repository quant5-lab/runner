package chartdata

import (
	"encoding/json"
	"math"
	"testing"
)

func TestPlotPoint_MarshalJSON_NaN(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: math.NaN(),
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// NaN should be encoded as null
	expected := `{"time":1234567890,"value":null}`
	if string(jsonBytes) != expected {
		t.Errorf("Expected %s, got %s", expected, string(jsonBytes))
	}

	// Verify it's valid JSON and value is null
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if result["value"] != nil {
		t.Errorf("Expected null value, got %v", result["value"])
	}
}

func TestPlotPoint_MarshalJSON_PositiveInf(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: math.Inf(1),
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// +Inf should be encoded as null
	expected := `{"time":1234567890,"value":null}`
	if string(jsonBytes) != expected {
		t.Errorf("Expected %s, got %s", expected, string(jsonBytes))
	}
}

func TestPlotPoint_MarshalJSON_NegativeInf(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: math.Inf(-1),
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// -Inf should be encoded as null
	expected := `{"time":1234567890,"value":null}`
	if string(jsonBytes) != expected {
		t.Errorf("Expected %s, got %s", expected, string(jsonBytes))
	}
}

func TestPlotPoint_MarshalJSON_Zero(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: 0.0,
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Zero should remain zero, not null
	expected := `{"time":1234567890,"value":0}`
	if string(jsonBytes) != expected {
		t.Errorf("Expected %s, got %s", expected, string(jsonBytes))
	}
}

func TestPlotPoint_MarshalJSON_NegativeZero(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: math.Copysign(0, -1), // -0.0
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// -0.0 should be encoded as 0 (JSON doesn't distinguish)
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if result["value"] == nil {
		t.Error("Expected numeric zero, got null")
	}
}

func TestPlotPoint_MarshalJSON_VerySmallNumber(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: 1e-308, // Very small but not zero
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should be encoded as number, not null
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if result["value"] == nil {
		t.Error("Expected number, got null")
	}
}

func TestPlotPoint_MarshalJSON_VeryLargeNumber(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: 1e308, // Very large but not infinity
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should be encoded as number, not null
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
	if result["value"] == nil {
		t.Error("Expected number, got null")
	}
}

func TestPlotPoint_MarshalJSON_NormalValues(t *testing.T) {
	testCases := []struct {
		name  string
		value float64
	}{
		{"positive", 123.456},
		{"negative", -123.456},
		{"integer", 42.0},
		{"fraction", 0.123456789},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			point := PlotPoint{
				Time:  1234567890,
				Value: tc.value,
			}

			jsonBytes, err := json.Marshal(point)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Should be valid JSON with numeric value
			var result map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &result); err != nil {
				t.Fatalf("Invalid JSON: %v", err)
			}
			if result["value"] == nil {
				t.Error("Expected numeric value, got null")
			}
		})
	}
}

func TestPlotPointSlice_MarshalJSON_Mixed(t *testing.T) {
	// Test slice with mixed valid/invalid values
	points := []PlotPoint{
		{Time: 1000, Value: math.NaN()},
		{Time: 2000, Value: 100.0},
		{Time: 3000, Value: math.Inf(1)},
		{Time: 4000, Value: 200.0},
		{Time: 5000, Value: math.Inf(-1)},
	}

	jsonBytes, err := json.Marshal(points)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify array structure
	var result []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if len(result) != 5 {
		t.Fatalf("Expected 5 points, got %d", len(result))
	}

	// Check NaN/Inf are null
	if result[0]["value"] != nil {
		t.Error("Point 0 (NaN) should be null")
	}
	if result[2]["value"] != nil {
		t.Error("Point 2 (+Inf) should be null")
	}
	if result[4]["value"] != nil {
		t.Error("Point 4 (-Inf) should be null")
	}

	// Check valid values are present
	if result[1]["value"] == nil {
		t.Error("Point 1 (100.0) should not be null")
	}
	if result[3]["value"] == nil {
		t.Error("Point 3 (200.0) should not be null")
	}
}

func TestPlotPoint_MarshalJSON_WithOptions(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: 123.45,
		Options: map[string]interface{}{
			"color": "red",
			"width": 2,
		},
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify valid JSON with options
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if result["value"] == nil {
		t.Error("Expected numeric value")
	}
	if result["options"] == nil {
		t.Error("Expected options to be present")
	}
}

func TestPlotPoint_MarshalJSON_NaNWithOptions(t *testing.T) {
	point := PlotPoint{
		Time:  1234567890,
		Value: math.NaN(),
		Options: map[string]interface{}{
			"pane": "indicator",
		},
	}

	jsonBytes, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify NaN becomes null but options preserved
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if result["value"] != nil {
		t.Error("Expected null value for NaN")
	}
	if result["options"] == nil {
		t.Error("Expected options to be preserved")
	}
}
