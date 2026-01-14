package security

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type ExpressionIdentifier interface {
	Identify(expr ast.Expression) string
}

type HashExpressionIdentifier struct{}

func NewHashExpressionIdentifier() *HashExpressionIdentifier {
	return &HashExpressionIdentifier{}
}

func (h *HashExpressionIdentifier) Identify(expr ast.Expression) string {
	data, _ := json.Marshal(expr)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:4])
}
