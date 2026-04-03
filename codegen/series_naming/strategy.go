package series_naming

import (
	"fmt"
)

/* Strategy defines interface for series variable naming approaches */
type Strategy interface {
	GenerateName(indicatorType string, periodPart string, sourceHash string) string
	GenerateDualPeriodName(indicatorType string, leftPeriod, rightPeriod int, sourceHash string) string
}

/* StatefulIndicatorNamer includes source hash for unique series per source expression */
type StatefulIndicatorNamer struct{}

func NewStatefulIndicatorNamer() *StatefulIndicatorNamer {
	return &StatefulIndicatorNamer{}
}

func (n *StatefulIndicatorNamer) GenerateName(indicatorType string, periodPart string, sourceHash string) string {
	if sourceHash == "" {
		return fmt.Sprintf("_%s_%s", indicatorType, periodPart)
	}
	return fmt.Sprintf("_%s_%s_%s", indicatorType, periodPart, sourceHash)
}

func (n *StatefulIndicatorNamer) GenerateDualPeriodName(indicatorType string, leftPeriod, rightPeriod int, sourceHash string) string {
	periodPart := fmt.Sprintf("%d_%d", leftPeriod, rightPeriod)
	if sourceHash == "" {
		return fmt.Sprintf("_%s_%s", indicatorType, periodPart)
	}
	return fmt.Sprintf("_%s_%s_%s", indicatorType, periodPart, sourceHash)
}

/* WindowBasedNamer excludes source hash - period alone determines window */
type WindowBasedNamer struct{}

func NewWindowBasedNamer() *WindowBasedNamer {
	return &WindowBasedNamer{}
}

func (n *WindowBasedNamer) GenerateName(indicatorType string, periodPart string, sourceHash string) string {
	return fmt.Sprintf("_%s_%s", indicatorType, periodPart)
}

func (n *WindowBasedNamer) GenerateDualPeriodName(indicatorType string, leftPeriod, rightPeriod int, sourceHash string) string {
	periodPart := fmt.Sprintf("%d_%d", leftPeriod, rightPeriod)
	return fmt.Sprintf("_%s_%s", indicatorType, periodPart)
}
