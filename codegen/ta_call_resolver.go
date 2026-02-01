package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type TACallResolver struct {
	registry *TASignatureRegistry
}

func NewTACallResolver(registry *TASignatureRegistry) *TACallResolver {
	return &TACallResolver{
		registry: registry,
	}
}

func (r *TACallResolver) Resolve(functionName string, call *ast.CallExpression) (*TAResolvedCall, error) {
	metadata, exists := r.registry.Lookup(functionName)
	if !exists {
		return r.resolveFallbackPattern(functionName, call)
	}

	providedArgCount := len(call.Arguments)
	overload, found := metadata.FindOverload(providedArgCount)
	if !found {
		return nil, fmt.Errorf(
			"function %s does not support %d arguments (supports: %d-%d)",
			functionName,
			providedArgCount,
			metadata.MinArgCount(),
			metadata.MaxArgCount(),
		)
	}

	resolved := NewTAResolvedCall(functionName)

	if overload.HasImplicitOHLC() {
		resolved.SetOHLCRequired()
	}

	needsDefaultSource := providedArgCount < metadata.MaxArgCount() && metadata.DefaultSource != ""

	if needsDefaultSource {
		resolved.ApplyDefaultSource(metadata.DefaultSource)
	}

	argumentIndex := 0
	for _, argSpec := range overload.Arguments {
		if argSpec.Classification == TAArgImplicitOHLC {
			continue
		}

		if argumentIndex >= len(call.Arguments) {
			break
		}

		arg := call.Arguments[argumentIndex]

		if argSpec.Classification.IsSeries() {
			resolved.AddSeriesArgument(arg)
		} else if argSpec.Classification.IsScalar() {
			resolved.AddScalarArgument(arg)
		}

		argumentIndex++
	}

	return resolved, nil
}

func (r *TACallResolver) resolveFallbackPattern(functionName string, call *ast.CallExpression) (*TAResolvedCall, error) {
	argCount := len(call.Arguments)
	if argCount != 2 {
		return nil, fmt.Errorf("unknown function %s requires exactly 2 arguments (source, length), got %d", functionName, argCount)
	}

	resolved := NewTAResolvedCall(functionName)
	resolved.AddSeriesArgument(call.Arguments[0])
	resolved.AddScalarArgument(call.Arguments[1])
	return resolved, nil
}
