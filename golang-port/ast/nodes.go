package ast

type NodeType string

const (
	TypeProgram             NodeType = "Program"
	TypeExpressionStatement NodeType = "ExpressionStatement"
	TypeCallExpression      NodeType = "CallExpression"
	TypeVariableDeclaration NodeType = "VariableDeclaration"
	TypeVariableDeclarator  NodeType = "VariableDeclarator"
	TypeMemberExpression    NodeType = "MemberExpression"
	TypeIdentifier          NodeType = "Identifier"
	TypeLiteral             NodeType = "Literal"
	TypeObjectExpression    NodeType = "ObjectExpression"
	TypeProperty            NodeType = "Property"
	TypeBinaryExpression    NodeType = "BinaryExpression"
	TypeIfStatement         NodeType = "IfStatement"
)

type Node interface {
	Type() NodeType
}

type Program struct {
	NodeType NodeType `json:"type"`
	Body     []Node   `json:"body"`
}

func (p *Program) Type() NodeType { return TypeProgram }

type ExpressionStatement struct {
	NodeType   NodeType   `json:"type"`
	Expression Expression `json:"expression"`
}

func (e *ExpressionStatement) Type() NodeType { return TypeExpressionStatement }

type Expression interface {
	Node
	expressionNode()
}

type CallExpression struct {
	NodeType  NodeType     `json:"type"`
	Callee    Expression   `json:"callee"`
	Arguments []Expression `json:"arguments"`
}

func (c *CallExpression) Type() NodeType  { return TypeCallExpression }
func (c *CallExpression) expressionNode() {}

type VariableDeclaration struct {
	NodeType     NodeType              `json:"type"`
	Declarations []VariableDeclarator  `json:"declarations"`
	Kind         string                `json:"kind"`
}

func (v *VariableDeclaration) Type() NodeType { return TypeVariableDeclaration }

type VariableDeclarator struct {
	NodeType NodeType   `json:"type"`
	ID       Identifier `json:"id"`
	Init     Expression `json:"init,omitempty"`
}

func (v *VariableDeclarator) Type() NodeType { return TypeVariableDeclarator }

type MemberExpression struct {
	NodeType NodeType   `json:"type"`
	Object   Expression `json:"object"`
	Property Expression `json:"property"`
	Computed bool       `json:"computed"`
}

func (m *MemberExpression) Type() NodeType  { return TypeMemberExpression }
func (m *MemberExpression) expressionNode() {}

type Identifier struct {
	NodeType NodeType `json:"type"`
	Name     string   `json:"name"`
}

func (i *Identifier) Type() NodeType  { return TypeIdentifier }
func (i *Identifier) expressionNode() {}

type Literal struct {
	NodeType NodeType    `json:"type"`
	Value    interface{} `json:"value"`
	Raw      string      `json:"raw"`
}

func (l *Literal) Type() NodeType  { return TypeLiteral }
func (l *Literal) expressionNode() {}

type ObjectExpression struct {
	NodeType   NodeType   `json:"type"`
	Properties []Property `json:"properties"`
}

func (o *ObjectExpression) Type() NodeType  { return TypeObjectExpression }
func (o *ObjectExpression) expressionNode() {}

type Property struct {
	NodeType  NodeType   `json:"type"`
	Key       Expression `json:"key"`
	Value     Expression `json:"value"`
	Kind      string     `json:"kind"`
	Method    bool       `json:"method"`
	Shorthand bool       `json:"shorthand"`
	Computed  bool       `json:"computed"`
}

func (p *Property) Type() NodeType { return TypeProperty }

type BinaryExpression struct {
	NodeType NodeType   `json:"type"`
	Operator string     `json:"operator"`
	Left     Expression `json:"left"`
	Right    Expression `json:"right"`
}

func (b *BinaryExpression) Type() NodeType  { return TypeBinaryExpression }
func (b *BinaryExpression) expressionNode() {}

type IfStatement struct {
	NodeType   NodeType   `json:"type"`
	Test       Expression `json:"test"`
	Consequent []Node     `json:"consequent"`
	Alternate  []Node     `json:"alternate,omitempty"`
}

func (i *IfStatement) Type() NodeType { return TypeIfStatement }
