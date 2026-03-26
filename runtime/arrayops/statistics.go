package arrayops

import (
	"math"
	"sort"

	"github.com/quant5-lab/runner/runtime/series"
)

type Statistics struct{}

func NewStatistics() *Statistics {
	return &Statistics{}
}

func (st *Statistics) Sum(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range slice {
		sum += v
	}
	return sum
}

func (st *Statistics) Avg(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	return st.Sum(s, offset) / float64(len(slice))
}

func (st *Statistics) Min(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func (st *Statistics) Max(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func (st *Statistics) Median(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	sorted := make([]float64, len(slice))
	copy(sorted, slice)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2.0
	}
	return sorted[mid]
}

func (st *Statistics) Mode(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	freq := make(map[float64]int)
	maxFreq := 0
	mode := slice[0]
	for _, v := range slice {
		freq[v]++
		if freq[v] > maxFreq {
			maxFreq = freq[v]
			mode = v
		}
	}
	return mode
}

func (st *Statistics) Stdev(s *series.ArraySeries, offset int, biased bool) float64 {
	variance := st.Variance(s, offset, biased)
	if math.IsNaN(variance) {
		return math.NaN()
	}
	return math.Sqrt(variance)
}

func (st *Statistics) Variance(s *series.ArraySeries, offset int, biased bool) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	if len(slice) == 1 {
		if biased {
			return 0.0
		}
		return math.NaN()
	}
	mean := st.Avg(s, offset)
	sumSq := 0.0
	for _, v := range slice {
		diff := v - mean
		sumSq += diff * diff
	}
	divisor := float64(len(slice))
	if !biased {
		divisor = float64(len(slice) - 1)
	}
	return sumSq / divisor
}

func (st *Statistics) Range(s *series.ArraySeries, offset int) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	return st.Max(s, offset) - st.Min(s, offset)
}

func (st *Statistics) Percentile(s *series.ArraySeries, offset int, percentile float64, method string) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	if percentile < 0 || percentile > 100 {
		return math.NaN()
	}
	sorted := make([]float64, len(slice))
	copy(sorted, slice)
	sort.Float64s(sorted)

	if method == "nearest_rank" {
		rank := int(math.Ceil(percentile / 100.0 * float64(len(sorted))))
		if rank == 0 {
			rank = 1
		}
		return sorted[rank-1]
	}

	pos := percentile / 100.0 * float64(len(sorted)-1)
	lower := int(math.Floor(pos))
	upper := int(math.Ceil(pos))
	if lower == upper {
		return sorted[lower]
	}
	weight := pos - float64(lower)
	return sorted[lower]*(1.0-weight) + sorted[upper]*weight
}

func (st *Statistics) PercentRank(s *series.ArraySeries, offset int, value float64) float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return math.NaN()
	}
	count := 0
	for _, v := range slice {
		if v < value {
			count++
		}
	}
	return float64(count) / float64(len(slice)) * 100.0
}

func (st *Statistics) Covariance(s1 *series.ArraySeries, offset1 int, s2 *series.ArraySeries, offset2 int, biased bool) float64 {
	slice1 := s1.Get(offset1)
	slice2 := s2.Get(offset2)
	if slice1 == nil || slice2 == nil || len(slice1) == 0 || len(slice2) == 0 {
		return math.NaN()
	}
	if len(slice1) != len(slice2) {
		return math.NaN()
	}
	if len(slice1) == 1 {
		if biased {
			return 0.0
		}
		return math.NaN()
	}
	stats := NewStatistics()
	mean1 := stats.Avg(s1, offset1)
	mean2 := stats.Avg(s2, offset2)
	sum := 0.0
	for i := 0; i < len(slice1); i++ {
		sum += (slice1[i] - mean1) * (slice2[i] - mean2)
	}
	divisor := float64(len(slice1))
	if !biased {
		divisor = float64(len(slice1) - 1)
	}
	return sum / divisor
}

func (st *Statistics) Standardize(s *series.ArraySeries, offset int) []float64 {
	slice := s.Get(offset)
	if len(slice) == 0 {
		return nil
	}
	mean := st.Avg(s, offset)
	stdev := st.Stdev(s, offset, false)
	if math.IsNaN(stdev) || stdev == 0.0 {
		result := make([]float64, len(slice))
		for i := range result {
			result[i] = 0.0
		}
		return result
	}
	result := make([]float64, len(slice))
	for i, v := range slice {
		result[i] = (v - mean) / stdev
	}
	return result
}

func (st *Statistics) Abs(s *series.ArraySeries, offset int) []float64 {
	slice := s.Get(offset)
	if slice == nil {
		return nil
	}
	result := make([]float64, len(slice))
	for i, v := range slice {
		result[i] = math.Abs(v)
	}
	return result
}
