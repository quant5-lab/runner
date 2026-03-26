package arrayops

func NewArrayWithValue(size int, value float64) []float64 {
	if size <= 0 {
		return []float64{}
	}
	arr := make([]float64, size)
	for i := range arr {
		arr[i] = value
	}
	return arr
}
