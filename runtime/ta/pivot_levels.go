package ta

import "math"

// PivotType identifies the pivot point calculation methodology.
type PivotType string

const (
	PivotTraditional PivotType = "Traditional"
	PivotFibonacci   PivotType = "Fibonacci"
	PivotWoodie      PivotType = "Woodie"
	PivotClassic     PivotType = "Classic"
	PivotDM          PivotType = "DM"
	PivotCamarilla   PivotType = "Camarilla"
)

// PivotLevelsSize is the fixed number of levels in every pivot_point_levels result.
// Index layout: [P, R1, S1, R2, S2, R3, S3, R4, S4, R5, S5]
const PivotLevelsSize = 11

// PivotLevelIndex maps semantic level names to their array positions.
const (
	PivotP  = 0
	PivotR1 = 1
	PivotS1 = 2
	PivotR2 = 3
	PivotS2 = 4
	PivotR3 = 5
	PivotS3 = 6
	PivotR4 = 7
	PivotS4 = 8
	PivotR5 = 9
	PivotS5 = 10
)

// NaNLevels returns a 11-element slice of NaN values, representing an
// unresolvable or pre-anchor pivot result.
func NaNLevels() []float64 {
	out := make([]float64, PivotLevelsSize)
	for i := range out {
		out[i] = math.NaN()
	}
	return out
}

// ComputePivotLevels calculates the 11 pivot levels for the given type using
// the completed period's OHLC values.
//
// Parameters:
//   - pt:               pivot methodology
//   - prevH/L/C/O:      completed period's High/Low/Close/Open
//   - woodieCurrentOpen: the new period's opening price (only relevant for Woodie)
//
// Returns NaN levels when prevH is NaN (no completed period available).
// Levels absent from a given type (e.g. R4/S4 for Fibonacci) return NaN.
func ComputePivotLevels(pt PivotType, prevH, prevL, prevC, prevO, woodieCurrentOpen float64) []float64 {
	if math.IsNaN(prevH) {
		return NaNLevels()
	}

	switch pt {
	case PivotTraditional:
		return computeTraditional(prevH, prevL, prevC)
	case PivotFibonacci:
		return computeFibonacci(prevH, prevL, prevC)
	case PivotWoodie:
		return computeWoodie(prevH, prevL, woodieCurrentOpen)
	case PivotClassic:
		return computeClassic(prevH, prevL, prevC)
	case PivotDM:
		return computeDM(prevH, prevL, prevC, prevO)
	case PivotCamarilla:
		return computeCamarilla(prevH, prevL, prevC)
	default:
		return NaNLevels()
	}
}

func computeTraditional(h, l, c float64) []float64 {
	out := make([]float64, PivotLevelsSize)
	p := (h + l + c) / 3
	hl := h - l
	out[PivotP] = p
	out[PivotR1] = 2*p - l
	out[PivotS1] = 2*p - h
	out[PivotR2] = p + hl
	out[PivotS2] = p - hl
	out[PivotR3] = h + 2*(p-l)
	out[PivotS3] = l - 2*(h-p)
	out[PivotR4] = h + 3*(p-l)
	out[PivotS4] = l - 3*(h-p)
	out[PivotR5] = h + 4*(p-l)
	out[PivotS5] = l - 4*(h-p)
	return out
}

func computeFibonacci(h, l, c float64) []float64 {
	out := NaNLevels()
	p := (h + l + c) / 3
	hl := h - l
	out[PivotP] = p
	out[PivotR1] = p + 0.382*hl
	out[PivotS1] = p - 0.382*hl
	out[PivotR2] = p + 0.618*hl
	out[PivotS2] = p - 0.618*hl
	out[PivotR3] = p + hl
	out[PivotS3] = p - hl
	return out
}

func computeWoodie(h, l, woodieOpen float64) []float64 {
	out := NaNLevels()
	p := (h + l + 2*woodieOpen) / 4
	hl := h - l
	out[PivotP] = p
	out[PivotR1] = 2*p - l
	out[PivotS1] = 2*p - h
	out[PivotR2] = p + hl
	out[PivotS2] = p - hl
	r3 := h + 2*(p-l)
	s3 := l - 2*(h-p)
	out[PivotR3] = r3
	out[PivotS3] = s3
	out[PivotR4] = r3 + hl
	out[PivotS4] = s3 - hl
	return out
}

func computeClassic(h, l, c float64) []float64 {
	out := NaNLevels()
	p := (h + l + c) / 3
	hl := h - l
	out[PivotP] = p
	out[PivotR1] = 2*p - l
	out[PivotS1] = 2*p - h
	out[PivotR2] = p + hl
	out[PivotS2] = p - hl
	out[PivotR3] = p + 2*hl
	out[PivotS3] = p - 2*hl
	out[PivotR4] = p + 3*hl
	out[PivotS4] = p - 3*hl
	return out
}

func computeDM(h, l, c, prevO float64) []float64 {
	out := NaNLevels()
	var x float64
	switch {
	case prevO == c:
		x = h + l + 2*c
	case c > prevO:
		x = 2*h + l + c
	default:
		x = 2*l + h + c
	}
	out[PivotP] = x / 4
	out[PivotR1] = x/2 - l
	out[PivotS1] = x/2 - h
	return out
}

func computeCamarilla(h, l, c float64) []float64 {
	out := make([]float64, PivotLevelsSize)
	p := (h + l + c) / 3
	hl := h - l
	out[PivotP] = p
	out[PivotR1] = c + 1.1*hl/12
	out[PivotS1] = c - 1.1*hl/12
	out[PivotR2] = c + 1.1*hl/6
	out[PivotS2] = c - 1.1*hl/6
	out[PivotR3] = c + 1.1*hl/4
	out[PivotS3] = c - 1.1*hl/4
	out[PivotR4] = c + 1.1*hl/2
	out[PivotS4] = c - 1.1*hl/2
	r5 := (h / l) * c
	out[PivotR5] = r5
	out[PivotS5] = c - (r5 - c)
	return out
}
