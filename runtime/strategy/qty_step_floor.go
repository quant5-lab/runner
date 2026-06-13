package strategy

import "math"

// floorToStep floors qty down to the nearest exact multiple of step.
// A step of zero or less signals an unknown lot size; qty is returned unchanged.
func floorToStep(qty, step float64) float64 {
	if step <= 0 {
		return qty
	}
	return math.Floor(qty/step) * step
}
