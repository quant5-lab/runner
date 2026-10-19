package strategy

import "math"

// intrabarSegment is a directed price move from start toward end.
type intrabarSegment struct {
	start float64
	end   float64
}

func intrabarSegments(barOpen, barHigh, barLow float64) [2]intrabarSegment {
	if IntrabarPath(barOpen, barHigh, barLow) == PathHighBeforeLow {
		return [2]intrabarSegment{{barOpen, barHigh}, {barHigh, barLow}}
	}
	return [2]intrabarSegment{{barOpen, barLow}, {barLow, barHigh}}
}

// contains reports whether level falls within this segment's price range (inclusive).
func (s intrabarSegment) contains(level float64) bool {
	lo, hi := math.Min(s.start, s.end), math.Max(s.start, s.end)
	return level >= lo && level <= hi
}

// proximityToStart returns the absolute distance from the segment's entry point to level.
func (s intrabarSegment) proximityToStart(level float64) float64 {
	return math.Abs(level - s.start)
}
