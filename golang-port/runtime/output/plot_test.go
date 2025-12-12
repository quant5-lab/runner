package output

import "testing"

func TestNewCollector(t *testing.T) {
	collector := NewCollector()
	if collector == nil {
		t.Fatal("NewCollector() returned nil")
	}
	if collector.series == nil {
		t.Error("Collector.series not initialized")
	}
}

func TestCollectorAdd(t *testing.T) {
	collector := NewCollector()

	collector.Add("SMA 20", 1000, 100.5, map[string]interface{}{"color": "#2962FF"})

	series := collector.GetSeries()
	if len(series) != 1 {
		t.Fatalf("Expected 1 series, got %d", len(series))
	}

	if series[0].Title != "SMA 20" {
		t.Errorf("Series title = %s, want SMA 20", series[0].Title)
	}

	if len(series[0].Data) != 1 {
		t.Fatalf("Expected 1 data point, got %d", len(series[0].Data))
	}

	point := series[0].Data[0]
	if point.Time != 1000 {
		t.Errorf("Point time = %d, want 1000", point.Time)
	}
	if point.Value != 100.5 {
		t.Errorf("Point value = %f, want 100.5", point.Value)
	}
	if point.Options["color"] != "#2962FF" {
		t.Errorf("Point color = %v, want #2962FF", point.Options["color"])
	}
}

func TestCollectorMultiplePoints(t *testing.T) {
	collector := NewCollector()

	collector.Add("Test", 1000, 100.0, nil)
	collector.Add("Test", 2000, 110.0, nil)
	collector.Add("Test", 3000, 105.0, nil)

	series := collector.GetSeries()
	if len(series) != 1 {
		t.Fatalf("Expected 1 series, got %d", len(series))
	}

	if len(series[0].Data) != 3 {
		t.Fatalf("Expected 3 data points, got %d", len(series[0].Data))
	}

	expectedValues := []float64{100.0, 110.0, 105.0}
	for i, point := range series[0].Data {
		if point.Value != expectedValues[i] {
			t.Errorf("Point[%d] value = %f, want %f", i, point.Value, expectedValues[i])
		}
	}
}

func TestCollectorMultipleSeries(t *testing.T) {
	collector := NewCollector()

	collector.Add("SMA 20", 1000, 100.0, nil)
	collector.Add("EMA 10", 1000, 102.0, nil)
	collector.Add("SMA 20", 2000, 105.0, nil)

	series := collector.GetSeries()
	if len(series) != 2 {
		t.Fatalf("Expected 2 series, got %d", len(series))
	}

	seriesByTitle := make(map[string]PlotSeries)
	for _, s := range series {
		seriesByTitle[s.Title] = s
	}

	if sma, ok := seriesByTitle["SMA 20"]; !ok {
		t.Error("SMA 20 series not found")
	} else if len(sma.Data) != 2 {
		t.Errorf("SMA 20 has %d points, want 2", len(sma.Data))
	}

	if ema, ok := seriesByTitle["EMA 10"]; !ok {
		t.Error("EMA 10 series not found")
	} else if len(ema.Data) != 1 {
		t.Errorf("EMA 10 has %d points, want 1", len(ema.Data))
	}
}

func TestCollectorEmptyOptions(t *testing.T) {
	collector := NewCollector()

	collector.Add("Test", 1000, 100.0, nil)

	series := collector.GetSeries()
	point := series[0].Data[0]

	if point.Options != nil {
		t.Errorf("Expected nil options, got %v", point.Options)
	}
}
