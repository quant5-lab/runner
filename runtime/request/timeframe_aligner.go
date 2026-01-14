package request

import "github.com/quant5-lab/runner/runtime/context"

type TimeframeAligner struct{}

func NewTimeframeAligner() *TimeframeAligner {
	return &TimeframeAligner{}
}

func (a *TimeframeAligner) FindSecurityBarIndex(
	securityContext *context.Context,
	mainTimeframeTimestamp int64,
	useCurrentBar bool,
) int {
	firstBarAfterCurrent := a.findFirstBarAfter(securityContext, mainTimeframeTimestamp)

	if firstBarAfterCurrent < 0 {
		return a.handleBeyondLastBar(securityContext, useCurrentBar)
	}

	return a.selectBarRelativeToFirst(firstBarAfterCurrent, useCurrentBar)
}

func (a *TimeframeAligner) findFirstBarAfter(
	securityContext *context.Context,
	timestamp int64,
) int {
	for i := 0; i < len(securityContext.Data); i++ {
		barTimestamp := securityContext.Data[i].Time

		if barTimestamp > timestamp {
			return i
		}
	}

	return -1
}

func (a *TimeframeAligner) handleBeyondLastBar(
	securityContext *context.Context,
	useCurrentBar bool,
) int {
	lastIndex := len(securityContext.Data) - 1

	if lastIndex < 0 {
		return -1
	}

	if useCurrentBar {
		return lastIndex
	}

	return lastIndex - 1
}

func (a *TimeframeAligner) selectBarRelativeToFirst(
	firstBarAfter int,
	useCurrentBar bool,
) int {
	if useCurrentBar {
		return firstBarAfter - 1
	}

	return firstBarAfter - 2
}
