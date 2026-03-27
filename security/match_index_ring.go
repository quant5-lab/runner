package security

// MatchIndexRing stores bar indices where a condition was true,
// retaining only the most recent N matches needed for occurrence-based lookback.
type MatchIndexRing struct {
	buffer   []int
	capacity int
	size     int
}

func NewMatchIndexRing(requiredOccurrence int) *MatchIndexRing {
	capacity := requiredOccurrence + 1
	return &MatchIndexRing{
		buffer:   make([]int, capacity),
		capacity: capacity,
		size:     0,
	}
}

func (r *MatchIndexRing) Push(barIndex int) {
	if r.size < r.capacity {
		r.buffer[r.size] = barIndex
		r.size++
	} else {
		copy(r.buffer, r.buffer[1:])
		r.buffer[r.capacity-1] = barIndex
	}
}

func (r *MatchIndexRing) GetNthMostRecent(n int) (int, bool) {
	if n < 0 || n >= r.size {
		return 0, false
	}
	idx := r.size - 1 - n
	return r.buffer[idx], true
}

func (r *MatchIndexRing) Size() int {
	return r.size
}
