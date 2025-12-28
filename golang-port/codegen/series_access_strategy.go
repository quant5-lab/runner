package codegen

import "fmt"

/* SeriesAccessStrategy abstracts series buffer access patterns across contexts.
 *
 * SRP: Single responsibility - generate series access code
 * OCP: Open for extension (new strategies), closed for modification
 * LSP: All strategies are substitutable
 * ISP: Minimal interface - only what's needed
 * DIP: Depend on abstraction, not concrete implementations
 */
type SeriesAccessStrategy interface {
	GenerateSet(varName string, valueExpr string) string
	GenerateGet(varName string, offset int) string
}

/* TopLevelSeriesAccessStrategy generates code for top-level series variables.
 *
 * Pattern: varNameSeries.Set(value), varNameSeries.Get(offset)
 * Used in: Main strategy loop, user-defined functions
 */
type TopLevelSeriesAccessStrategy struct{}

func NewTopLevelSeriesAccessStrategy() *TopLevelSeriesAccessStrategy {
	return &TopLevelSeriesAccessStrategy{}
}

func (s *TopLevelSeriesAccessStrategy) GenerateSet(varName string, valueExpr string) string {
	return fmt.Sprintf("%sSeries.Set(%s)", varName, valueExpr)
}

func (s *TopLevelSeriesAccessStrategy) GenerateGet(varName string, offset int) string {
	return fmt.Sprintf("%sSeries.Get(%d)", varName, offset)
}

/* ArrowContextSeriesAccessStrategy generates code for arrow function context.
 *
 * Pattern: arrowCtx.GetOrCreateSeries("varName").Set(value)
 * Used in: Arrow function inline IIFEs
 */
type ArrowContextSeriesAccessStrategy struct{}

func NewArrowContextSeriesAccessStrategy() *ArrowContextSeriesAccessStrategy {
	return &ArrowContextSeriesAccessStrategy{}
}

func (s *ArrowContextSeriesAccessStrategy) GenerateSet(varName string, valueExpr string) string {
	return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Set(%s)", varName, valueExpr)
}

func (s *ArrowContextSeriesAccessStrategy) GenerateGet(varName string, offset int) string {
	return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(%d)", varName, offset)
}
