package preprocessor

import (
	"strings"

	"github.com/quant5-lab/runner/parser"
)

func qualifiedIdentToMemberAccess(qualified string) (*parser.MemberAccess, bool) {
	dot := strings.Index(qualified, ".")
	if dot < 0 {
		return nil, false
	}
	parts := strings.SplitN(qualified, ".", 2)
	props := strings.Split(parts[1], ".")
	return &parser.MemberAccess{Object: parts[0], Properties: props}, true
}

func rewriteFactorIdent(factor *parser.Factor, mappings map[string]string, shadow map[string]bool) {
	if factor == nil || factor.Ident == nil {
		return
	}
	name := *factor.Ident
	if shadow[name] {
		return
	}
	qualified, ok := mappings[name]
	if !ok {
		return
	}
	ma, ok := qualifiedIdentToMemberAccess(qualified)
	if !ok {
		return
	}
	factor.Ident = nil
	factor.MemberAccess = ma
}

func rewritePrimaryExprIdent(primary *parser.PrimaryExpr, mappings map[string]string, shadow map[string]bool) {
	if primary == nil || primary.Ident == nil {
		return
	}
	name := *primary.Ident
	if shadow[name] {
		return
	}
	qualified, ok := mappings[name]
	if !ok {
		return
	}
	ma, ok := qualifiedIdentToMemberAccess(qualified)
	if !ok {
		return
	}
	primary.Ident = nil
	primary.MemberAccess = ma
}
