package math

import (
	gomath "math"
	"testing"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"Positive", 42.5, 42.5},
		{"Negative", -42.5, 42.5},
		{"Zero", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Abs(tt.input)
			if got != tt.want {
				t.Errorf("Abs(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"Two values", []float64{10.0, 20.0}, 20.0},
		{"Three values", []float64{10.0, 30.0, 20.0}, 30.0},
		{"Negative values", []float64{-10.0, -5.0, -20.0}, -5.0},
		{"Single value", []float64{42.0}, 42.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Max(tt.values...)
			if got != tt.want {
				t.Errorf("Max(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestMaxEmpty(t *testing.T) {
	got := Max()
	if !gomath.IsNaN(got) {
		t.Errorf("Max() = %v, want NaN", got)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"Two values", []float64{10.0, 20.0}, 10.0},
		{"Three values", []float64{10.0, 30.0, 5.0}, 5.0},
		{"Negative values", []float64{-10.0, -5.0, -20.0}, -20.0},
		{"Single value", []float64{42.0}, 42.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Min(tt.values...)
			if got != tt.want {
				t.Errorf("Min(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestMinEmpty(t *testing.T) {
	got := Min()
	if !gomath.IsNaN(got) {
		t.Errorf("Min() = %v, want NaN", got)
	}
}

func TestPow(t *testing.T) {
	tests := []struct {
		name string
		x, y float64
		want float64
	}{
		{"2^3", 2.0, 3.0, 8.0},
		{"10^2", 10.0, 2.0, 100.0},
		{"5^0", 5.0, 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Pow(tt.x, tt.y)
			if got != tt.want {
				t.Errorf("Pow(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"Perfect square", 16.0, 4.0},
		{"Non-perfect", 2.0, gomath.Sqrt(2.0)},
		{"Zero", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sqrt(tt.input)
			if gomath.Abs(got-tt.want) > 1e-10 {
				t.Errorf("Sqrt(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFloor(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"Positive decimal", 42.7, 42.0},
		{"Negative decimal", -42.7, -43.0},
		{"Integer", 10.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Floor(tt.input)
			if got != tt.want {
				t.Errorf("Floor(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCeil(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"Positive decimal", 42.3, 43.0},
		{"Negative decimal", -42.3, -42.0},
		{"Integer", 10.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Ceil(tt.input)
			if got != tt.want {
				t.Errorf("Ceil(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"Round up", 42.6, 43.0},
		{"Round down", 42.4, 42.0},
		{"Exact half", 42.5, 43.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Round(tt.input)
			if got != tt.want {
				t.Errorf("Round(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLog(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"e^1", gomath.E, 1.0},
		{"e^2", gomath.E * gomath.E, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Log(tt.input)
			if gomath.Abs(got-tt.want) > 1e-10 {
				t.Errorf("Log(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExp(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"e^0", 0.0, 1.0},
		{"e^1", 1.0, gomath.E},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Exp(tt.input)
			if gomath.Abs(got-tt.want) > 1e-10 {
				t.Errorf("Exp(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"Positive values", []float64{1.0, 2.0, 3.0}, 6.0},
		{"Mixed values", []float64{10.0, -5.0, 2.5}, 7.5},
		{"Empty slice", []float64{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sum(tt.values)
			if got != tt.want {
				t.Errorf("Sum(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestAvg(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"Three values", []float64{10.0, 20.0, 30.0}, 20.0},
		{"Two values", []float64{5.0, 15.0}, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Avg(tt.values...)
			if got != tt.want {
				t.Errorf("Avg(%v) = %v, want %v", tt.values, got, tt.want)
			}
		})
	}
}

func TestAvgEmpty(t *testing.T) {
	got := Avg()
	if !gomath.IsNaN(got) {
		t.Errorf("Avg() = %v, want NaN", got)
	}
}
