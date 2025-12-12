package codegen

import "fmt"

type CodeTransformer interface {
	Transform(code string) string
}

type addNotEqualZeroTransformer struct{}

func (t *addNotEqualZeroTransformer) Transform(code string) string {
	return fmt.Sprintf("%s != 0", code)
}

type addParenthesesTransformer struct{}

func (t *addParenthesesTransformer) Transform(code string) string {
	return fmt.Sprintf("(%s)", code)
}

func NewAddNotEqualZeroTransformer() CodeTransformer {
	return &addNotEqualZeroTransformer{}
}

func NewAddParenthesesTransformer() CodeTransformer {
	return &addParenthesesTransformer{}
}
