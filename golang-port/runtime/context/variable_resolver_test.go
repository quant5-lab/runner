package context

import (
	"testing"

	"github.com/quant5-lab/runner/runtime/series"
)

func TestRecursiveResolver_LocalVariableFound(t *testing.T) {
	localRegistry := NewMapBasedRegistry()
	testSeries := series.NewSeries(10)
	testSeries.Set(42.0)
	localRegistry.Set("localVar", testSeries)

	resolver := NewRecursiveResolver(localRegistry, NewIdentityAligner(), nil)

	result := resolver.Resolve("localVar", 0)

	if !result.Found {
		t.Fatal("expected variable to be found")
	}
	if result.Series != testSeries {
		t.Error("expected same series instance")
	}
}

func TestRecursiveResolver_ParentVariableFound(t *testing.T) {
	parentRegistry := NewMapBasedRegistry()
	parentSeries := series.NewSeries(10)
	parentSeries.Set(100.0)
	parentRegistry.Set("parentVar", parentSeries)
	parentResolver := NewRecursiveResolver(parentRegistry, NewIdentityAligner(), nil)

	childRegistry := NewMapBasedRegistry()
	childResolver := NewRecursiveResolver(childRegistry, NewIdentityAligner(), parentResolver)

	result := childResolver.Resolve("parentVar", 0)

	if !result.Found {
		t.Fatal("expected parent variable to be found")
	}
	if result.Series != parentSeries {
		t.Error("expected parent series instance")
	}
}

func TestRecursiveResolver_GrandparentVariableFound(t *testing.T) {
	grandparentRegistry := NewMapBasedRegistry()
	grandparentSeries := series.NewSeries(10)
	grandparentRegistry.Set("grandparentVar", grandparentSeries)
	grandparentResolver := NewRecursiveResolver(grandparentRegistry, NewIdentityAligner(), nil)

	parentRegistry := NewMapBasedRegistry()
	parentResolver := NewRecursiveResolver(parentRegistry, NewIdentityAligner(), grandparentResolver)

	childRegistry := NewMapBasedRegistry()
	childResolver := NewRecursiveResolver(childRegistry, NewIdentityAligner(), parentResolver)

	result := childResolver.Resolve("grandparentVar", 0)

	if !result.Found {
		t.Fatal("expected grandparent variable to be found")
	}
	if result.Series != grandparentSeries {
		t.Error("expected grandparent series instance")
	}
}

func TestRecursiveResolver_VariableNotFound(t *testing.T) {
	resolver := NewRecursiveResolver(NewMapBasedRegistry(), NewIdentityAligner(), nil)

	result := resolver.Resolve("nonexistent", 0)

	if result.Found {
		t.Error("expected variable not to be found")
	}
}

func TestRecursiveResolver_BarIndexAlignment(t *testing.T) {
	parentRegistry := NewMapBasedRegistry()
	parentSeries := series.NewSeries(100)
	for i := 0; i < 10; i++ {
		parentSeries.Set(float64(i * 10))
	}
	parentRegistry.Set("parentVar", parentSeries)
	parentResolver := NewRecursiveResolver(parentRegistry, NewIdentityAligner(), nil)

	aligner := NewMappedAligner()
	aligner.SetMapping(0, 5)
	aligner.SetMapping(1, 10)
	aligner.SetMapping(2, 15)

	childRegistry := NewMapBasedRegistry()
	childResolver := NewRecursiveResolver(childRegistry, aligner, parentResolver)

	result := childResolver.Resolve("parentVar", 1)

	if !result.Found {
		t.Fatal("expected variable to be found")
	}
	if result.SourceBarIdx != 10 {
		t.Errorf("expected aligned bar index 10, got %d", result.SourceBarIdx)
	}
}

func TestRecursiveResolver_LocalVariableShadowsParent(t *testing.T) {
	parentRegistry := NewMapBasedRegistry()
	parentSeries := series.NewSeries(10)
	parentSeries.Set(100.0)
	parentRegistry.Set("sharedVar", parentSeries)
	parentResolver := NewRecursiveResolver(parentRegistry, NewIdentityAligner(), nil)

	childRegistry := NewMapBasedRegistry()
	childSeries := series.NewSeries(10)
	childSeries.Set(200.0)
	childRegistry.Set("sharedVar", childSeries)
	childResolver := NewRecursiveResolver(childRegistry, NewIdentityAligner(), parentResolver)

	result := childResolver.Resolve("sharedVar", 0)

	if !result.Found {
		t.Fatal("expected variable to be found")
	}
	if result.Series != childSeries {
		t.Error("expected child series to shadow parent")
	}
}
