// Package codegen provides PineScript to Go code generation.
//
// # Strategy Direction Resolution
//
// Extracts strategy direction parameters from PineScript AST expressions
// using Chain of Responsibility pattern.
//
// Architecture:
//
//	DirectionExtractor (interface)
//	├── MemberExpressionDirectionExtractor     (v5: strategy.long/short)
//	├── BooleanLiteralDirectionExtractor        (v4: true/false literals)
//	├── VariableDirectionExtractor              (runtime string variable)
//	├── IdentifierDirectionExtractor            (v4: "true"/"false" identifiers)
//	└── contextualConditionalDirectionExtractor (ternary: cond ? strategy.long : strategy.short)
//
// Production usage (generator-bound, handles all expression forms):
//
//	extractor := NewContextAwareDirectionExtractor(gen)
//	direction := extractor.Extract(callArg)
//
// Version Compatibility:
//
//	v4: strategy.entry("id", true)                                 → BooleanLiteralDirectionExtractor
//	v5: strategy.entry("id", strategy.long)                        → MemberExpressionDirectionExtractor
//	v5: strategy.entry("id", bull ? strategy.long : strategy.short) → contextualConditionalDirectionExtractor
//
// Extensibility:
//
// Add new PineScript version support:
//  1. Implement DirectionExtractor interface
//  2. Register in NewContextAwareDirectionExtractor chain
//  3. Existing extractors unchanged (Open/Closed Principle)
package codegen
