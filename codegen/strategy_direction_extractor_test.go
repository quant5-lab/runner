package codegen

import (
	"testing"

	"github.com/quant5-lab/runner/ast"
)

func newComponentChain() *ChainDirectionExtractor {
	return NewChainDirectionExtractor(
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&IdentifierDirectionExtractor{},
	)
}

func newNamedArgChain() *ChainDirectionExtractor {
	chain := &ChainDirectionExtractor{}
	chain.extractors = []DirectionExtractor{
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&IdentifierDirectionExtractor{},
		&namedDirectionArgExtractor{chain: chain},
	}
	return chain
}

func makeNamedArgObj(keyName string, value ast.Expression) *ast.ObjectExpression {
	return &ast.ObjectExpression{
		NodeType: ast.TypeObjectExpression,
		Properties: []ast.Property{
			{Key: &ast.Identifier{Name: keyName}, Value: value},
		},
	}
}

func TestMemberExpressionDirectionExtractor_TwoLevelForm(t *testing.T) {
	e := &MemberExpressionDirectionExtractor{}

	tests := []struct {
		name      string
		obj       string
		prop      string
		wantDir   string
		wantFound bool
	}{
		{"strategy.long resolves", "strategy", "long", "strategy.Long", true},
		{"strategy.short resolves", "strategy", "short", "strategy.Short", true},
		{"any object with long resolves", "position", "long", "strategy.Long", true},
		{"unknown property rejected", "strategy", "unknown", "", false},
		{"case-sensitive long vs Long", "strategy", "Long", "", false},
		{"case-sensitive short vs Short", "strategy", "Short", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &ast.MemberExpression{
				Object:   &ast.Identifier{Name: tt.obj},
				Property: &ast.Identifier{Name: tt.prop},
			}
			dir, found := e.Extract(expr)
			if found != tt.wantFound {
				t.Errorf("found=%v, want %v", found, tt.wantFound)
			}
			if dir != tt.wantDir {
				t.Errorf("dir=%q, want %q", dir, tt.wantDir)
			}
		})
	}
}

func TestMemberExpressionDirectionExtractor_StrategyDirectionNamespace(t *testing.T) {
	e := &MemberExpressionDirectionExtractor{}

	directionExpr := func(prop string) ast.Expression {
		return &ast.MemberExpression{
			Object: &ast.MemberExpression{
				Object:   &ast.Identifier{Name: "strategy"},
				Property: &ast.Identifier{Name: "direction"},
			},
			Property: &ast.Identifier{Name: prop},
		}
	}

	tests := []struct {
		name      string
		expr      ast.Expression
		wantDir   string
		wantFound bool
	}{
		{"strategy.direction.long", directionExpr("long"), "strategy.DirectionLong", true},
		{"strategy.direction.short", directionExpr("short"), "strategy.DirectionShort", true},
		{"strategy.direction.all", directionExpr("all"), "strategy.DirectionAll", true},
		{"strategy.direction.unknown rejected", directionExpr("unknown"), "", false},
		{"strategy.direction.LONG (wrong case) rejected", directionExpr("LONG"), "", false},
		{"strategy.direction.Short (wrong case) rejected", directionExpr("Short"), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, found := e.Extract(tt.expr)
			if found != tt.wantFound {
				t.Errorf("found=%v, want %v", found, tt.wantFound)
			}
			if dir != tt.wantDir {
				t.Errorf("dir=%q, want %q", dir, tt.wantDir)
			}
		})
	}
}

func TestMemberExpressionDirectionExtractor_TypeSafety(t *testing.T) {
	e := &MemberExpressionDirectionExtractor{}

	tests := []struct {
		name  string
		expr  ast.Expression
		found bool
	}{
		{"literal rejected", &ast.Literal{Value: true}, false},
		{"identifier rejected", &ast.Identifier{Name: "long"}, false},
		{"nil property rejected", &ast.MemberExpression{
			Object: &ast.Identifier{Name: "strategy"}, Property: nil,
		}, false},
		{"non-identifier property rejected", &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Literal{Value: "long"},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := e.Extract(tt.expr)
			if found != tt.found {
				t.Errorf("found=%v, want %v", found, tt.found)
			}
		})
	}
}

func TestBooleanLiteralDirectionExtractor_Mapping(t *testing.T) {
	e := &BooleanLiteralDirectionExtractor{}

	tests := []struct {
		name      string
		value     interface{}
		wantDir   string
		wantFound bool
	}{
		{"true → Long", true, "strategy.Long", true},
		{"false → Short", false, "strategy.Short", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, found := e.Extract(&ast.Literal{Value: tt.value})
			if found != tt.wantFound {
				t.Errorf("found=%v, want %v", found, tt.wantFound)
			}
			if dir != tt.wantDir {
				t.Errorf("dir=%q, want %q", dir, tt.wantDir)
			}
		})
	}
}

func TestBooleanLiteralDirectionExtractor_TypeDiscrimination(t *testing.T) {
	e := &BooleanLiteralDirectionExtractor{}

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{"string literal", &ast.Literal{Value: "true"}},
		{"integer literal", &ast.Literal{Value: 1}},
		{"float literal", &ast.Literal{Value: 1.0}},
		{"nil literal value", &ast.Literal{Value: nil}},
		{"identifier", &ast.Identifier{Name: "true"}},
		{"member expression", &ast.MemberExpression{Object: &ast.Identifier{Name: "x"}, Property: &ast.Identifier{Name: "y"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := e.Extract(tt.expr)
			if found {
				t.Error("want found=false, got true")
			}
		})
	}
}

func TestIdentifierDirectionExtractor_PineV4BareNames(t *testing.T) {
	e := &IdentifierDirectionExtractor{}

	tests := []struct {
		name      string
		identName string
		wantDir   string
		wantFound bool
	}{
		{"true → Long", "true", "strategy.Long", true},
		{"false → Short", "false", "strategy.Short", true},
		{"True (wrong case) rejected", "True", "", false},
		{"FALSE (wrong case) rejected", "FALSE", "", false},
		{"True mixed-case rejected", "tRuE", "", false},
		{"arbitrary identifier rejected", "longDirection", "", false},
		{"empty identifier rejected", "", "", false},
		{"whitespace-padded rejected", " true ", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, found := e.Extract(&ast.Identifier{Name: tt.identName})
			if found != tt.wantFound {
				t.Errorf("found=%v, want %v", found, tt.wantFound)
			}
			if dir != tt.wantDir {
				t.Errorf("dir=%q, want %q", dir, tt.wantDir)
			}
		})
	}
}

func TestIdentifierDirectionExtractor_NonIdentifierRejected(t *testing.T) {
	e := &IdentifierDirectionExtractor{}

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{"string literal", &ast.Literal{Value: "true"}},
		{"bool literal", &ast.Literal{Value: true}},
		{"member expression", &ast.MemberExpression{
			Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found := e.Extract(tt.expr)
			if found {
				t.Errorf("%s: want not found, got found", tt.name)
			}
		})
	}
}

func TestVariableDirectionExtractor(t *testing.T) {
	tests := []struct {
		name      string
		vars      map[string]string
		expr      ast.Expression
		wantDir   string
		wantFound bool
	}{
		{
			"registered string variable resolves as-is",
			map[string]string{"entryDir": "string"},
			&ast.Identifier{Name: "entryDir"},
			"entryDir", true,
		},
		{
			"non-string variable rejected",
			map[string]string{"count": "int"},
			&ast.Identifier{Name: "count"},
			"", false,
		},
		{
			"float variable rejected",
			map[string]string{"level": "float"},
			&ast.Identifier{Name: "level"},
			"", false,
		},
		{
			"unregistered identifier rejected",
			map[string]string{},
			&ast.Identifier{Name: "unknownDir"},
			"", false,
		},
		{
			"non-identifier expression rejected",
			map[string]string{"x": "string"},
			&ast.Literal{Value: "x"},
			"", false,
		},
		{
			"nil expression rejected",
			map[string]string{"x": "string"},
			nil,
			"", false,
		},
		{
			"empty identifier name rejected even when registered",
			map[string]string{"": "string"},
			&ast.Identifier{Name: ""},
			"", false,
		},
		{
			"nil variable map treats any identifier as unregistered",
			nil,
			&ast.Identifier{Name: "dir"},
			"", false,
		},
		{
			"empty-string type value is not a string-typed variable",
			map[string]string{"dir": ""},
			&ast.Identifier{Name: "dir"},
			"", false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &VariableDirectionExtractor{variables: tt.vars}
			dir, found := e.Extract(tt.expr)
			if found != tt.wantFound {
				t.Errorf("found=%v, want %v", found, tt.wantFound)
			}
			if dir != tt.wantDir {
				t.Errorf("dir=%q, want %q", dir, tt.wantDir)
			}
		})
	}
}

func TestChainDirectionExtractor_Resolve_ResolvableExpressions(t *testing.T) {
	chain := newComponentChain()

	tests := []struct {
		name    string
		expr    ast.Expression
		wantDir string
		pineVer string
	}{
		{
			name:    "strategy.long member",
			pineVer: "v5",
			expr:    &ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
			wantDir: "strategy.Long",
		},
		{
			name:    "strategy.short member",
			pineVer: "v5",
			expr:    &ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "short"}},
			wantDir: "strategy.Short",
		},
		{name: "true literal", pineVer: "v4", expr: &ast.Literal{Value: true}, wantDir: "strategy.Long"},
		{name: "false literal", pineVer: "v4", expr: &ast.Literal{Value: false}, wantDir: "strategy.Short"},
		{name: "true identifier", pineVer: "v4", expr: &ast.Identifier{Name: "true"}, wantDir: "strategy.Long"},
		{name: "false identifier", pineVer: "v4", expr: &ast.Identifier{Name: "false"}, wantDir: "strategy.Short"},
	}

	for _, tt := range tests {
		t.Run(tt.pineVer+"/"+tt.name, func(t *testing.T) {
			dir, err := chain.Resolve(tt.expr)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if dir != tt.wantDir {
				t.Errorf("want %q, got %q", tt.wantDir, dir)
			}
		})
	}
}

func TestChainDirectionExtractor_Resolve_UnresolvableExpressions(t *testing.T) {
	chain := newComponentChain()

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{"numeric literal", &ast.Literal{Value: 42.0}},
		{"string literal", &ast.Literal{Value: "unknown"}},
		{"nil literal value", &ast.Literal{Value: nil}},
		{"call expression", &ast.CallExpression{Callee: &ast.Identifier{Name: "getDirection"}}},
		{"unary expression", &ast.UnaryExpression{Operator: "!", Argument: &ast.Identifier{Name: "x"}}},
		{"binary expression", &ast.BinaryExpression{Left: &ast.Identifier{Name: "a"}, Operator: "+", Right: &ast.Literal{Value: 1.0}}},
		{"arbitrary identifier", &ast.Identifier{Name: "myDirection"}},
		{"nil expression", nil},
		{"ObjectExpression without long property", &ast.ObjectExpression{
			NodeType:   ast.TypeObjectExpression,
			Properties: []ast.Property{{Key: &ast.Identifier{Name: "qty"}, Value: &ast.Literal{Value: 1.0}}},
		}},
		{"empty ObjectExpression", &ast.ObjectExpression{NodeType: ast.TypeObjectExpression}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := chain.Resolve(tc.expr)
			if err == nil {
				t.Errorf("want error for %T, got direction %q", tc.expr, dir)
			}
			if dir != "" {
				t.Errorf("want empty string on error for %T, got %q", tc.expr, dir)
			}
		})
	}
}

func TestChainDirectionExtractor_ChainOrdering(t *testing.T) {
	chain := newComponentChain()

	tests := []struct {
		name    string
		expr    ast.Expression
		wantDir string
	}{
		{
			"member expression resolved by MemberExtractor (not Identifier)",
			&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
			"strategy.Long",
		},
		{
			"bool literal resolved by BooleanExtractor (not Identifier)",
			&ast.Literal{Value: true},
			"strategy.Long",
		},
		{
			"true identifier resolved by IdentifierExtractor",
			&ast.Identifier{Name: "true"},
			"strategy.Long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := chain.Resolve(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if dir != tt.wantDir {
				t.Errorf("want %q, got %q", tt.wantDir, dir)
			}
		})
	}
}

// TestChainDirectionExtractor_ExtensibilityContract proves that new extractors
// can be composed into the chain without modifying existing ones, and that
// earlier extractors in the chain continue to resolve their own shapes correctly
// when the chain is extended.
func TestChainDirectionExtractor_ExtensibilityContract(t *testing.T) {
	base := NewChainDirectionExtractor(
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
	)

	dir, err := base.Resolve(&ast.Literal{Value: true})
	if err != nil || dir != "strategy.Long" {
		t.Errorf("base chain Bool literal: want (strategy.Long, nil), got (%q, %v)", dir, err)
	}
	_, err = base.Resolve(&ast.Identifier{Name: "true"})
	if err == nil {
		t.Error("base chain without IdentifierExtractor must not resolve bare identifier")
	}

	extended := NewChainDirectionExtractor(
		&MemberExpressionDirectionExtractor{},
		&BooleanLiteralDirectionExtractor{},
		&IdentifierDirectionExtractor{},
	)
	dir, err = extended.Resolve(&ast.Literal{Value: false})
	if err != nil || dir != "strategy.Short" {
		t.Errorf("extended chain Bool literal: want (strategy.Short, nil), got (%q, %v)", dir, err)
	}
	dir, err = extended.Resolve(&ast.Identifier{Name: "true"})
	if err != nil || dir != "strategy.Long" {
		t.Errorf("extended chain Identifier: want (strategy.Long, nil), got (%q, %v)", dir, err)
	}
}

func TestChainDirectionExtractor_Resolve_VariableResolution(t *testing.T) {
	gen := newTestGenerator()
	gen.variables["entryDir"] = "string"

	chain := NewContextAwareDirectionExtractor(gen)

	dir, err := chain.Resolve(&ast.Identifier{Name: "entryDir"})
	if err != nil {
		t.Fatalf("unexpected error for registered string variable: %v", err)
	}
	if dir != "entryDir" {
		t.Errorf("want %q, got %q", "entryDir", dir)
	}
}

func TestChainDirectionExtractor_Resolve_UnregisteredVariableIsError(t *testing.T) {
	gen := newTestGenerator()
	chain := NewContextAwareDirectionExtractor(gen)

	_, err := chain.Resolve(&ast.Identifier{Name: "unknownVar"})
	if err == nil {
		t.Error("want error for unregistered identifier, got nil")
	}
}

func TestContextAwareDirectionExtractor_ConditionalExpression(t *testing.T) {
	const pine = `
//@version=5
strategy("Direction Test")
sma_bullish = close > ta.sma(close, 20)
entry_type = sma_bullish ? strategy.long : strategy.short
if barstate.isconfirmed
    strategy.entry("E", entry_type)
`
	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	if !contains(code, "strategy.Long") && !contains(code, "strategy.Short") {
		t.Errorf("generated code must reference strategy.Long or strategy.Short:\n%s", code)
	}
	if contains(code, `"strategy.long"`) || contains(code, `"strategy.short"`) {
		t.Errorf("direction must not be emitted as a string literal:\n%s", code)
	}
}

func TestContextAwareDirectionExtractor_MemberExpressionDirect(t *testing.T) {
	const pine = `
//@version=5
strategy("Long Only")
if close > ta.sma(close, 20)
    strategy.entry("Long", strategy.long)
`
	code, err := compilePineScript(pine)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	if !contains(code, "strategy.Long") {
		t.Errorf("expected strategy.Long in generated code:\n%s", code)
	}
	if contains(code, "strategy.Short") {
		t.Errorf("unexpected strategy.Short in long-only strategy:\n%s", code)
	}
}

// TestContextAwareDirectionExtractor_ConditionalBranchResolvesViaFullChain proves
// that ternary branch resolution delegates to chain.Resolve — meaning any extractor
// registered in the chain (including extractors absent from newComponentChain) is
// automatically available for branch resolution.  This is the extensibility contract
// for conditional direction expressions.
func TestContextAwareDirectionExtractor_ConditionalBranchResolvesViaFullChain(t *testing.T) {
	gen := newTestGenerator()
	gen.literalFormatter = NewLiteralFormatter()

	gen.variables["dirVar"] = "string"

	chain := NewContextAwareDirectionExtractor(gen)

	expr := &ast.ConditionalExpression{
		Test:       &ast.Literal{Value: true},
		Consequent: &ast.Identifier{Name: "dirVar"},
		Alternate: &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: "short"},
		},
	}

	dir, err := chain.Resolve(expr)
	if err != nil {
		t.Fatalf("unexpected error resolving ternary with variable branch: %v", err)
	}
	if !contains(dir, "dirVar") {
		t.Errorf("resolved direction must reference the variable branch, got: %s", dir)
	}
	if !contains(dir, "strategy.Short") {
		t.Errorf("resolved direction must reference strategy.Short for alternate branch, got: %s", dir)
	}
}

func TestNamedDirectionArgExtractor_DirectionPropertyMapping(t *testing.T) {
	chain := newNamedArgChain()
	e := &namedDirectionArgExtractor{chain: chain}

	longShortMember := func(dir string) *ast.MemberExpression {
		return &ast.MemberExpression{
			Object:   &ast.Identifier{Name: "strategy"},
			Property: &ast.Identifier{Name: dir},
		}
	}

	tests := []struct {
		name       string
		properties []ast.Property
		wantDir    string
	}{
		{
			"long=strategy.short → Short (sole property)",
			[]ast.Property{{Key: &ast.Identifier{Name: "long"}, Value: longShortMember("short")}},
			"strategy.Short",
		},
		{
			"long=strategy.long → Long (sole property)",
			[]ast.Property{{Key: &ast.Identifier{Name: "long"}, Value: longShortMember("long")}},
			"strategy.Long",
		},
		{
			"long=true literal → Long",
			[]ast.Property{{Key: &ast.Identifier{Name: "long"}, Value: &ast.Literal{Value: true}}},
			"strategy.Long",
		},
		{
			"long=false literal → Short",
			[]ast.Property{{Key: &ast.Identifier{Name: "long"}, Value: &ast.Literal{Value: false}}},
			"strategy.Short",
		},
		{
			"long=identifier true → Long",
			[]ast.Property{{Key: &ast.Identifier{Name: "long"}, Value: &ast.Identifier{Name: "true"}}},
			"strategy.Long",
		},
		{
			"long=identifier false → Short",
			[]ast.Property{{Key: &ast.Identifier{Name: "long"}, Value: &ast.Identifier{Name: "false"}}},
			"strategy.Short",
		},
		{
			"long= preceded by other properties",
			[]ast.Property{
				{Key: &ast.Identifier{Name: "qty"}, Value: &ast.Literal{Value: 10.0}},
				{Key: &ast.Identifier{Name: "long"}, Value: longShortMember("short")},
			},
			"strategy.Short",
		},
		{
			"long= surrounded by qty/comment/when properties",
			[]ast.Property{
				{Key: &ast.Identifier{Name: "qty"}, Value: &ast.Literal{Value: 1.0}},
				{Key: &ast.Identifier{Name: "long"}, Value: longShortMember("long")},
				{Key: &ast.Identifier{Name: "comment"}, Value: &ast.Literal{Value: "buy"}},
				{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "cond"}},
			},
			"strategy.Long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := &ast.ObjectExpression{NodeType: ast.TypeObjectExpression, Properties: tt.properties}
			dir, found := e.Extract(obj)
			if !found {
				t.Fatalf("want found=true, got false")
			}
			if dir != tt.wantDir {
				t.Errorf("got %q, want %q", dir, tt.wantDir)
			}
		})
	}
}

// TestNamedDirectionArgExtractor_TypeGate guards against panics on unexpected
// input types and missing/unresolvable "long" properties.
func TestNamedDirectionArgExtractor_TypeGate(t *testing.T) {
	chain := newNamedArgChain()
	e := &namedDirectionArgExtractor{chain: chain}

	tests := []struct {
		name string
		expr ast.Expression
	}{
		{
			"member expression rejected",
			&ast.MemberExpression{Object: &ast.Identifier{Name: "strategy"}, Property: &ast.Identifier{Name: "long"}},
		},
		{"bool literal rejected", &ast.Literal{Value: true}},
		{"identifier rejected", &ast.Identifier{Name: "true"}},
		{"nil expression rejected", nil},
		{"empty ObjectExpression yields not-found", &ast.ObjectExpression{NodeType: ast.TypeObjectExpression}},
		{
			"ObjectExpression without long property yields not-found",
			&ast.ObjectExpression{
				NodeType: ast.TypeObjectExpression,
				Properties: []ast.Property{
					{Key: &ast.Identifier{Name: "qty"}, Value: &ast.Literal{Value: 100.0}},
					{Key: &ast.Identifier{Name: "when"}, Value: &ast.Identifier{Name: "cond"}},
				},
			},
		},
		{
			"ObjectExpression with non-identifier key skipped, yields not-found",
			&ast.ObjectExpression{
				NodeType: ast.TypeObjectExpression,
				Properties: []ast.Property{
					{Key: &ast.Literal{Value: "long"}, Value: &ast.MemberExpression{
						Object:   &ast.Identifier{Name: "strategy"},
						Property: &ast.Identifier{Name: "short"},
					}},
				},
			},
		},
		{"long value unresolvable by chain yields not-found", makeNamedArgObj("long", &ast.Literal{Value: "invalid_string"})},
		{"long value numeric literal yields not-found", makeNamedArgObj("long", &ast.Literal{Value: 1.0})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, found := e.Extract(tt.expr)
			if found {
				t.Errorf("want found=false, got true (dir=%q)", dir)
			}
			if dir != "" {
				t.Errorf("want empty dir on not-found, got %q", dir)
			}
		})
	}
}

// TestChainDirectionExtractor_NamedArgDelegatesValueThroughChain proves that
// namedDirectionArgExtractor passes the "long" property value to chain.Resolve,
// enabling extractors only present in NewContextAwareDirectionExtractor (absent from
// newNamedArgChain) to resolve it — the same delegation pattern used by
// contextualConditionalDirectionExtractor for ternary branches.
func TestChainDirectionExtractor_NamedArgDelegatesValueThroughChain(t *testing.T) {
	gen := newTestGenerator()
	gen.variables["entryDir"] = "string"
	chain := NewContextAwareDirectionExtractor(gen)

	obj := makeNamedArgObj("long", &ast.Identifier{Name: "entryDir"})
	dir, err := chain.Resolve(obj)
	if err != nil {
		t.Fatalf("namedDirectionArgExtractor must delegate long= value to full chain: %v", err)
	}
	if dir != "entryDir" {
		t.Errorf("got %q, want %q", dir, "entryDir")
	}
}
