package codegen

// StatefulIndicatorContext defines execution context for stateful indicators (RMA, EMA).
// Separates concerns: context knowledge vs calculation logic.
type StatefulIndicatorContext interface {
	// GenerateSeriesAccess returns code to access the series buffer for reading
	GenerateSeriesAccess(varName string, offset int) string

	// GenerateSeriesUpdate returns code to update the series buffer
	GenerateSeriesUpdate(varName string, value string) string

	// IsWithinArrowFunction returns true if generating code within arrow function scope
	IsWithinArrowFunction() bool
}

// TopLevelIndicatorContext generates code for indicators in main execution scope
type TopLevelIndicatorContext struct{}

func NewTopLevelIndicatorContext() *TopLevelIndicatorContext {
	return &TopLevelIndicatorContext{}
}

func (c *TopLevelIndicatorContext) GenerateSeriesAccess(varName string, offset int) string {
	return formatSeriesGet(varName, offset)
}

func (c *TopLevelIndicatorContext) GenerateSeriesUpdate(varName string, value string) string {
	return formatSeriesSet(varName, value)
}

func (c *TopLevelIndicatorContext) IsWithinArrowFunction() bool {
	return false
}

// ArrowFunctionIndicatorContext generates code for indicators within arrow functions
type ArrowFunctionIndicatorContext struct{}

func NewArrowFunctionIndicatorContext() *ArrowFunctionIndicatorContext {
	return &ArrowFunctionIndicatorContext{}
}

func (c *ArrowFunctionIndicatorContext) GenerateSeriesAccess(varName string, offset int) string {
	return formatArrowSeriesGet(varName, offset)
}

func (c *ArrowFunctionIndicatorContext) GenerateSeriesUpdate(varName string, value string) string {
	return formatArrowSeriesSet(varName, value)
}

func (c *ArrowFunctionIndicatorContext) IsWithinArrowFunction() bool {
	return true
}
