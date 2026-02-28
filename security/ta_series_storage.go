package security

import (
	"math"

	"github.com/quant5-lab/runner/runtime/series"
)

type TASeriesStorage interface {
	Set(barIndex int, value float64)
	Get(barIndex int) float64
}

type ScalarStorage struct {
	lastValue float64
	lastBar   int
}

func NewScalarStorage() *ScalarStorage {
	return &ScalarStorage{
		lastBar: -1,
	}
}

func (s *ScalarStorage) Set(barIndex int, value float64) {
	s.lastValue = value
	s.lastBar = barIndex
}

func (s *ScalarStorage) Get(barIndex int) float64 {
	if barIndex == s.lastBar {
		return s.lastValue
	}
	return math.NaN()
}

type SeriesStorage struct {
	buffer *series.Series
}

func NewSeriesStorage(capacity int) *SeriesStorage {
	if capacity <= 0 {
		capacity = 1
	}
	return &SeriesStorage{
		buffer: series.NewSeries(capacity),
	}
}

func (s *SeriesStorage) Set(barIndex int, value float64) {
	currentPosition := s.buffer.Position()
	if barIndex == currentPosition {
		s.buffer.Set(value)
	} else if barIndex == currentPosition+1 {
		s.buffer.Next()
		s.buffer.Set(value)
	}
}

func (s *SeriesStorage) Get(barIndex int) float64 {
	currentPosition := s.buffer.Position()
	if barIndex > currentPosition {
		return math.NaN()
	}
	offset := currentPosition - barIndex
	return s.buffer.Get(offset)
}
