package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

type nestedTACallEntry struct {
	seriesName string
	preamble   string
}

// nestedTACallPrecomputer resolves the outer-TA source-at-offset problem: when a binary
// or conditional expression used as a TA source itself contains TA calls, those inner
// calls must be evaluated bar-by-bar before the outer TA loop so their history is
// available via arrowCtx series. Identical calls are deduplicated by expression hash.
type nestedTACallPrecomputer struct {
	exprGenerator ArrowExpressionGenerator
	entries       map[string]nestedTACallEntry
}

func newNestedTACallPrecomputer(exprGen ArrowExpressionGenerator) *nestedTACallPrecomputer {
	return &nestedTACallPrecomputer{
		exprGenerator: exprGen,
		entries:       make(map[string]nestedTACallEntry),
	}
}

func (p *nestedTACallPrecomputer) Process(expr ast.Expression) error {
	return p.walk(expr)
}

func (p *nestedTACallPrecomputer) walk(expr ast.Expression) error {
	switch e := expr.(type) {
	case *ast.CallExpression:
		funcName := extractCallFunctionName(e)
		if isTAFunction(funcName) {
			return p.precompute(e)
		}
		for _, arg := range e.Arguments {
			if err := p.walk(arg); err != nil {
				return err
			}
		}
	case *ast.BinaryExpression:
		if err := p.walk(e.Left); err != nil {
			return err
		}
		return p.walk(e.Right)
	case *ast.UnaryExpression:
		return p.walk(e.Argument)
	case *ast.ConditionalExpression:
		if err := p.walk(e.Test); err != nil {
			return err
		}
		if err := p.walk(e.Consequent); err != nil {
			return err
		}
		return p.walk(e.Alternate)
	}
	return nil
}

func (p *nestedTACallPrecomputer) precompute(call *ast.CallExpression) error {
	hasher := &ExpressionHasher{}
	hash := hasher.Hash(call)
	if _, exists := p.entries[hash]; exists {
		return nil
	}

	callCode, err := p.exprGenerator.Generate(call)
	if err != nil {
		return fmt.Errorf("precompute nested TA %s: %w", extractCallFunctionName(call), err)
	}

	seriesName := "_ta_src_" + hash
	preamble := fmt.Sprintf(
		"%sSeries := arrowCtx.GetOrCreateSeries(%q); %sSeries.Set(%s)",
		seriesName, seriesName, seriesName, callCode,
	)
	p.entries[hash] = nestedTACallEntry{seriesName: seriesName, preamble: preamble}
	return nil
}

func (p *nestedTACallPrecomputer) BuildLookup() CallVarLookup {
	hasher := &ExpressionHasher{}
	return func(call *ast.CallExpression) string {
		hash := hasher.Hash(call)
		if entry, ok := p.entries[hash]; ok {
			return entry.seriesName
		}
		return ""
	}
}

func (p *nestedTACallPrecomputer) Preamble() string {
	if len(p.entries) == 0 {
		return ""
	}
	parts := make([]string, 0, len(p.entries))
	for _, entry := range p.entries {
		parts = append(parts, entry.preamble)
	}
	return strings.Join(parts, "; ")
}

func (p *nestedTACallPrecomputer) HasNestedCalls() bool {
	return len(p.entries) > 0
}
