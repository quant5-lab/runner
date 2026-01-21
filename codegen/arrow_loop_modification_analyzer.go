package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

/*
ArrowLoopModificationAnalyzer detects variables modified inside for-loops.

When a variable is modified inside a for-loop via Series.Set(),
the return statement must use varSeries.GetCurrent() instead of
the scalar variable, because the scalar is not updated in loops.

Example Pine:

	count = 0
	for i = 0 to len - 1
	    count := count + 1   // Generates: countSeries.Set(countSeries.GetCurrent() + 1)
	count                    // Return must be: countSeries.GetCurrent(), not count

This analyzer walks the AST to find:
1. Variables declared BEFORE a for-loop
2. That are REASSIGNED inside the for-loop body
*/
type ArrowLoopModificationAnalyzer struct{}

/* NewArrowLoopModificationAnalyzer creates analyzer instance */
func NewArrowLoopModificationAnalyzer() *ArrowLoopModificationAnalyzer {
	return &ArrowLoopModificationAnalyzer{}
}

/* FindLoopModifiedVariables returns set of variable names modified in for-loops
 *
 * Universal Series Algorithm:
 * 1. Collect ALL variables declared anywhere in function (including loop bodies)
 * 2. For each for-loop, track which variables exist BEFORE entering that loop
 * 3. Mark variables as loop-modified if they're reassigned inside the loop AND existed before
 * 4. Variables declared inside a loop and only used there are loop-local (not tracked)
 */
func (a *ArrowLoopModificationAnalyzer) FindLoopModifiedVariables(body []ast.Node) map[string]bool {
	result := make(map[string]bool)

	// Track all variables and where they're declared
	a.analyzeWithScope(body, make(map[string]bool), result)

	return result
}

/* analyzeWithScope recursively analyzes loops with proper scope tracking
 *
 * declaredBefore: variables that exist before entering current scope
 * modified: output set of loop-modified variables
 */
func (a *ArrowLoopModificationAnalyzer) analyzeWithScope(statements []ast.Node, declaredBefore map[string]bool, modified map[string]bool) {
	// Track variables declared at this scope level
	currentScope := make(map[string]bool)
	for k, v := range declaredBefore {
		currentScope[k] = v
	}

	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.VariableDeclaration:
			// Add new declarations to current scope
			for _, declarator := range s.Declarations {
				if id, ok := declarator.ID.(*ast.Identifier); ok {
					currentScope[id.Name] = true
				} else if arrayPattern, ok := declarator.ID.(*ast.ArrayPattern); ok {
					for _, elem := range arrayPattern.Elements {
						currentScope[elem.Name] = true
					}
				}
			}

		case *ast.ForStatement:
			// Analyze loop body with current scope as "declared before"
			a.analyzeForLoopWithScope(s.Body, currentScope, modified)

		case *ast.IfStatement:
			// Recurse into if-statement branches
			a.analyzeWithScope(s.Consequent, currentScope, modified)
			a.analyzeWithScope(s.Alternate, currentScope, modified)
		}
	}
}

/* analyzeForLoopWithScope finds reassignments inside for-loop body */
func (a *ArrowLoopModificationAnalyzer) analyzeForLoopWithScope(body []ast.Node, declaredBefore map[string]bool, modified map[string]bool) {
	for _, stmt := range body {
		switch s := stmt.(type) {
		case *ast.VariableDeclaration:
			// Check if this is a reassignment (variable existed before loop)
			for _, declarator := range s.Declarations {
				if id, ok := declarator.ID.(*ast.Identifier); ok {
					if declaredBefore[id.Name] {
						modified[id.Name] = true
					}
				}
			}

		case *ast.IfStatement:
			// Recurse into if-statements inside loop
			a.analyzeForLoopWithScope(s.Consequent, declaredBefore, modified)
			a.analyzeForLoopWithScope(s.Alternate, declaredBefore, modified)

		case *ast.ForStatement:
			// Nested loop: build scope including variables declared in outer loop
			nestedScope := make(map[string]bool)
			for k, v := range declaredBefore {
				nestedScope[k] = v
			}

			// Add variables declared in current loop body (before nested loop)
			for _, preStmt := range body {
				if preStmt == s {
					break // Stop before nested loop
				}
				if varDecl, ok := preStmt.(*ast.VariableDeclaration); ok {
					for _, declarator := range varDecl.Declarations {
						if id, ok := declarator.ID.(*ast.Identifier); ok {
							nestedScope[id.Name] = true
						} else if arrayPattern, ok := declarator.ID.(*ast.ArrayPattern); ok {
							for _, elem := range arrayPattern.Elements {
								nestedScope[elem.Name] = true
							}
						}
					}
				}
			}

			// Analyze nested loop with extended scope
			a.analyzeForLoopWithScope(s.Body, nestedScope, modified)
		}
	}
}
