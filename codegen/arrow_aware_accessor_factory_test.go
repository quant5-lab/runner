package codegen

import (
	"strings"
	"testing"

	"github.com/quant5-lab/runner/ast"
)

/* TestArrowAwareAccessorFactory_TaTrMemberExpression validates ta.tr MemberExpression in arrow context */
func TestArrowAwareAccessorFactory_TaTrMemberExpression(t *testing.T) {
	gen := newTestGenerator()
	accessResolver := NewArrowSeriesAccessResolver()
	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	exprGen := &legacyArrowExpressionGenerator{gen: gen}

	factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

	expr := &ast.MemberExpression{
		Object:   &ast.Identifier{Name: "ta"},
		Property: &ast.Identifier{Name: "tr"},
		Computed: false,
	}

	accessor, err := factory.CreateAccessorForExpression(expr)
	if err != nil {
		t.Fatalf("Failed to create accessor for ta.tr: %v", err)
	}

	if accessor == nil {
		t.Fatal("CreateAccessorForExpression(ta.tr) returned nil accessor")
	}

	// Verify accessor type
	trAccessor, ok := accessor.(*BuiltinTrueRangeAccessor)
	if !ok {
		t.Fatalf("Expected *BuiltinTrueRangeAccessor, got %T", accessor)
	}

	// Verify current value access
	currentCode := trAccessor.GenerateCurrentValueAccess()
	if !strings.Contains(currentCode, "ctx.BarIndex") {
		t.Errorf("GenerateCurrentValueAccess() should contain ctx.BarIndex, got: %s", currentCode)
	}

	if !strings.Contains(currentCode, "math.Max") {
		t.Errorf("GenerateCurrentValueAccess() should contain math.Max for tr calculation, got: %s", currentCode)
	}

	// Verify loop value access
	loopCode := trAccessor.GenerateLoopValueAccess("j")
	if !strings.Contains(loopCode, "ctx.BarIndex-j") {
		t.Errorf("GenerateLoopValueAccess() should contain ctx.BarIndex-j, got: %s", loopCode)
	}

	if !strings.Contains(loopCode, "math.Max") {
		t.Errorf("GenerateLoopValueAccess() should contain math.Max for tr calculation, got: %s", loopCode)
	}

	// Verify no Series.Get() in arrow context
	if strings.Contains(loopCode, "Series.Get(") {
		t.Errorf("GenerateLoopValueAccess() should not use Series.Get() in arrow context, got: %s", loopCode)
	}
}

/* TestArrowAwareAccessorFactory_TrIdentifier validates bare tr identifier in arrow context */
func TestArrowAwareAccessorFactory_TrIdentifier(t *testing.T) {
	gen := newTestGenerator()
	accessResolver := NewArrowSeriesAccessResolver()
	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	exprGen := &legacyArrowExpressionGenerator{gen: gen}

	factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

	trIdentifier := &ast.Identifier{Name: "tr"}

	accessor, err := factory.CreateAccessorForExpression(trIdentifier)
	if err != nil {
		t.Fatalf("Failed to create accessor for tr: %v", err)
	}

	if accessor == nil {
		t.Fatal("CreateAccessorForExpression(tr) returned nil accessor")
	}

	// Verify it returns BuiltinTrueRangeAccessor
	if _, ok := accessor.(*BuiltinTrueRangeAccessor); !ok {
		t.Fatalf("Expected *BuiltinTrueRangeAccessor for tr identifier, got %T", accessor)
	}
}

/* TestArrowAwareAccessorFactory_TrNotConfusedWithParameter validates tr builtin vs parameter */
func TestArrowAwareAccessorFactory_TrNotConfusedWithParameter(t *testing.T) {
	gen := newTestGenerator()
	gen.variables = map[string]string{"my_tr": "float"} // User variable named my_tr
	accessResolver := NewArrowSeriesAccessResolver()
	identifierResolver := NewArrowIdentifierResolver(accessResolver)
	exprGen := &legacyArrowExpressionGenerator{gen: gen}

	factory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)

	t.Run("tr builtin", func(t *testing.T) {
		trIdentifier := &ast.Identifier{Name: "tr"}
		accessor, err := factory.CreateAccessorForExpression(trIdentifier)

		if err != nil {
			t.Fatalf("CreateAccessorForExpression(tr) error: %v", err)
		}

		if _, ok := accessor.(*BuiltinTrueRangeAccessor); !ok {
			t.Errorf("tr should return BuiltinTrueRangeAccessor, got %T", accessor)
		}
	})

	t.Run("my_tr parameter", func(t *testing.T) {
		myTrIdentifier := &ast.Identifier{Name: "my_tr"}
		accessor, err := factory.CreateAccessorForExpression(myTrIdentifier)

		if err != nil {
			t.Fatalf("CreateAccessorForExpression(my_tr) error: %v", err)
		}

		// Should NOT be BuiltinTrueRangeAccessor
		if _, ok := accessor.(*BuiltinTrueRangeAccessor); ok {
			t.Errorf("my_tr parameter should not return BuiltinTrueRangeAccessor, got %T", accessor)
		}
	})
}
