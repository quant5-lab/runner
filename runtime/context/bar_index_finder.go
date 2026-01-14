package context

type BarIndexFinder struct{}

func NewBarIndexFinder() *BarIndexFinder {
	return &BarIndexFinder{}
}

func (f *BarIndexFinder) FindContainingBar(data []OHLCV, targetTimestamp int64) int {
	if len(data) == 0 {
		return -1
	}

	firstBarAfter := f.findFirstBarAfter(data, targetTimestamp)

	if firstBarAfter < 0 {
		return f.handleBeyondLastBar(data)
	}

	return f.selectBarBeforeBoundary(firstBarAfter)
}

func (f *BarIndexFinder) findFirstBarAfter(data []OHLCV, timestamp int64) int {
	for i := 0; i < len(data); i++ {
		if data[i].Time > timestamp {
			return i
		}
	}
	return -1
}

func (f *BarIndexFinder) handleBeyondLastBar(data []OHLCV) int {
	return len(data) - 1
}

func (f *BarIndexFinder) selectBarBeforeBoundary(boundaryIndex int) int {
	if boundaryIndex > 0 {
		return boundaryIndex - 1
	}
	return -1
}
