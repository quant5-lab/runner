package ast

type NodeType string

const (
	TypeProgram                 NodeType = "Program"
	TypeExpressionStatement     NodeType = "ExpressionStatement"
	TypeCallExpression          NodeType = "CallExpression"
	TypeVariableDeclaration     NodeType = "VariableDeclaration"
	TypeVariableDeclarator      NodeType = "VariableDeclarator"
	TypeMemberExpression        NodeType = "MemberExpression"
	TypeIdentifier              NodeType = "Identifier"
	TypeLiteral                 NodeType = "Literal"
	TypeObjectExpression        NodeType = "ObjectExpression"
	TypeProperty                NodeType = "Property"
	TypeBinaryExpression        NodeType = "BinaryExpression"
	TypeIfStatement             NodeType = "IfStatement"
	TypeForStatement            NodeType = "ForStatement"
	TypeForInStatement          NodeType = "ForInStatement"
	TypeWhileStatement          NodeType = "WhileStatement"
	TypeConditionalExpression   NodeType = "ConditionalExpression"
	TypeLogicalExpression       NodeType = "LogicalExpression"
	TypeUnaryExpression         NodeType = "UnaryExpression"
	TypeArrayPattern            NodeType = "ArrayPattern"
	TypeArrowFunctionExpression NodeType = "ArrowFunctionExpression"
	TypeBreakStatement          NodeType = "BreakStatement"
	TypeContinueStatement       NodeType = "ContinueStatement"
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
	NodeType     NodeType             `json:"type"`
	Declarations []VariableDeclarator `json:"declarations"`
	Kind         string               `json:"kind"`
}

func (v *VariableDeclaration) Type() NodeType { return TypeVariableDeclaration }

type VariableDeclarator struct {
	NodeType NodeType   `json:"type"`
	ID       Pattern    `json:"id"`
	Init     Expression `json:"init,omitempty"`
}

func (v *VariableDeclarator) Type() NodeType { return TypeVariableDeclarator }

type Pattern interface {
	Node
	patternNode()
}

type ArrayPattern struct {
	NodeType NodeType     `json:"type"`
	Elements []Identifier `json:"elements"`
}

func (a *ArrayPattern) Type() NodeType { return TypeArrayPattern }
func (a *ArrayPattern) patternNode()   {}

func (i *Identifier) patternNode() {}

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

func (i *IfStatement) Type() NodeType  { return TypeIfStatement }
func (i *IfStatement) expressionNode() {}

type ForStatement struct {
	NodeType NodeType   `json:"type"`
	Counter  string     `json:"counter"`
	From     Expression `json:"from"`
	To       Expression `json:"to"`
	Step     Expression `json:"step,omitempty"`
	Body     []Node     `json:"body"`
}

func (f *ForStatement) Type() NodeType  { return TypeForStatement }
func (f *ForStatement) expressionNode() {}

type ForInStatement struct {
	NodeType   NodeType   `json:"type"`
	IndexVar   string     `json:"indexVar,omitempty"`
	ElementVar string     `json:"elementVar"`
	Collection Expression `json:"collection"`
	Body       []Node     `json:"body"`
}

func (f *ForInStatement) Type() NodeType  { return TypeForInStatement }
func (f *ForInStatement) expressionNode() {}

type WhileStatement struct {
	NodeType  NodeType   `json:"type"`
	Condition Expression `json:"condition"`
	Body      []Node     `json:"body"`
}

func (w *WhileStatement) Type() NodeType  { return TypeWhileStatement }
func (w *WhileStatement) expressionNode() {}

type ConditionalExpression struct {
	NodeType   NodeType   `json:"type"`
	Test       Expression `json:"test"`
	Consequent Expression `json:"consequent"`
	Alternate  Expression `json:"alternate"`
}

func (c *ConditionalExpression) Type() NodeType  { return TypeConditionalExpression }
func (c *ConditionalExpression) expressionNode() {}

type LogicalExpression struct {
	NodeType NodeType   `json:"type"`
	Operator string     `json:"operator"`
	Left     Expression `json:"left"`
	Right    Expression `json:"right"`
}

func (l *LogicalExpression) Type() NodeType  { return TypeLogicalExpression }
func (l *LogicalExpression) expressionNode() {}

type UnaryExpression struct {
	NodeType NodeType   `json:"type"`
	Operator string     `json:"operator"`
	Argument Expression `json:"argument"`
	Prefix   bool       `json:"prefix"`
}

func (u *UnaryExpression) Type() NodeType  { return TypeUnaryExpression }
func (u *UnaryExpression) expressionNode() {}

type ArrowFunctionExpression struct {
	NodeType NodeType     `json:"type"`
	Params   []Identifier `json:"params"`
	Body     []Node       `json:"body"`
}

func (a *ArrowFunctionExpression) Type() NodeType  { return TypeArrowFunctionExpression }
func (a *ArrowFunctionExpression) expressionNode() {}

type BreakStatement struct {
	NodeType NodeType `json:"type"`
}

func (b *BreakStatement) Type() NodeType { return TypeBreakStatement }

type ContinueStatement struct {
	NodeType NodeType `json:"type"`
}

func (c *ContinueStatement) Type() NodeType { return TypeContinueStatement }
