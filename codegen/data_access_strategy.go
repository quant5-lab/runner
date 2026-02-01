package codegen

import "fmt"

// DataAccessStrategy defines how to generate code for accessing series data.
type DataAccessStrategy interface {
	GenerateInitialValueAccess(period int) string
	GenerateLoopValueAccess(loopVar string) string
	GenerateCurrentValueAccess() string
}

// SeriesDataAccessor generates code for accessing user-defined Series variables.
type SeriesDataAccessor struct {
	variableName string
	offset       HistoricalOffset
}

// NewSeriesDataAccessor creates accessor for Series variable.
func NewSeriesDataAccessor(variableName string, offset HistoricalOffset) *SeriesDataAccessor {
	return &SeriesDataAccessor{
		variableName: variableName,
		offset:       offset,
	}
}

func (a *SeriesDataAccessor) GenerateInitialValueAccess(period int) string {
	totalOffset := a.offset.Add(period - 1)
	return fmt.Sprintf("%sSeries.Get(%d)", a.variableName, totalOffset)
}

func (a *SeriesDataAccessor) GenerateLoopValueAccess(loopVar string) string {
	accessExpr := a.offset.FormatLoopAccess(loopVar)
	return fmt.Sprintf("%sSeries.Get(%s)", a.variableName, accessExpr)
}

func (a *SeriesDataAccessor) GenerateCurrentValueAccess() string {
	if a.offset.IsZero() {
		return fmt.Sprintf("%sSeries.GetCurrent()", a.variableName)
	}
	return fmt.Sprintf("%sSeries.Get(%d)", a.variableName, a.offset.Value())
}

// OHLCVDataAccessor generates code for accessing built-in OHLCV fields.
type OHLCVDataAccessor struct {
	fieldName string
	offset    HistoricalOffset
}

// NewOHLCVDataAccessor creates accessor for OHLCV field.
func NewOHLCVDataAccessor(fieldName string, offset HistoricalOffset) *OHLCVDataAccessor {
	return &OHLCVDataAccessor{
		fieldName: fieldName,
		offset:    offset,
	}
}

func (a *OHLCVDataAccessor) GenerateInitialValueAccess(period int) string {
	totalOffset := a.offset.Add(period - 1)
	seriesName := OHLCVFieldToSeriesName(a.fieldName)
	return fmt.Sprintf("%s.Get(%d)", seriesName, totalOffset)
}

func (a *OHLCVDataAccessor) GenerateLoopValueAccess(loopVar string) string {
	seriesName := OHLCVFieldToSeriesName(a.fieldName)
	if a.offset.IsZero() {
		return fmt.Sprintf("%s.Get(%s)", seriesName, loopVar)
	}
	return fmt.Sprintf("%s.Get(%s+%d)", seriesName, loopVar, a.offset.Value())
}

func (a *OHLCVDataAccessor) GenerateCurrentValueAccess() string {
	seriesName := OHLCVFieldToSeriesName(a.fieldName)
	if a.offset.IsZero() {
		return fmt.Sprintf("%s.GetCurrent()", seriesName)
	}
	return fmt.Sprintf("%s.Get(%d)", seriesName, a.offset.Value())
}

// DataAccessFactory creates appropriate accessor based on source classification.
type DataAccessFactory struct{}

// CreateAccessor returns the correct DataAccessStrategy for the given source.
func (f *DataAccessFactory) CreateAccessor(source SourceInfo) DataAccessStrategy {
	offset := NewHistoricalOffset(source.BaseOffset)

	if source.IsSeriesVariable() {
		return NewSeriesDataAccessor(source.VariableName, offset)
	}
	return NewOHLCVDataAccessor(source.FieldName, offset)
}
