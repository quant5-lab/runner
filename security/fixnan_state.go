package security

import "math"

type FixnanState struct {
	lastValidValue float64
}

func NewFixnanState() *FixnanState {
	return &FixnanState{
		lastValidValue: math.NaN(),
	}
}

func (s *FixnanState) ForwardFill(value float64) float64 {
	if !math.IsNaN(value) {
		s.lastValidValue = value
		return value
	}
	return s.lastValidValue
}
