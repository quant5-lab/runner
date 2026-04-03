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
//	├── MemberExpressionDirectionExtractor  (v5: strategy.long/short)
//	├── BooleanLiteralDirectionExtractor     (v4: true/false literals)
//	└── IdentifierDirectionExtractor         (v4: "true"/"false" identifiers)
//
// Usage:
//
//	extractor := NewDefaultDirectionExtractor()
//	direction := extractor.Extract(callArg)
//
// Version Compatibility:
//
//	v4: strategy.entry("id", true)             → BooleanLiteralDirectionExtractor
//	v5: strategy.entry("id", strategy.long)    → MemberExpressionDirectionExtractor
//
// Extensibility:
//
// Add new PineScript version support:
//  1. Implement DirectionExtractor interface
//  2. Add to NewDefaultDirectionExtractor() chain
//  3. Existing code unchanged (Open/Closed Principle)
package codegen
