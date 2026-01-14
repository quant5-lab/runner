package pivot

type ComparisonType int

const (
	GreaterThan ComparisonType = iota
	LessThan
)

type Window struct {
	leftBars  int
	rightBars int
}

func NewWindow(leftBars, rightBars int) Window {
	return Window{
		leftBars:  leftBars,
		rightBars: rightBars,
	}
}

func (w Window) TotalWidth() int {
	return w.leftBars + 1 + w.rightBars
}

func (w Window) CenterOffset() int {
	return w.rightBars
}
