package codegen

/* CodeGenContext encapsulates code generation state for tuple indicators */
type CodeGenContext struct {
	Indenter       func() string
	IndentLevel    *int
	IncreaseIndent func()
	DecreaseIndent func()
}
