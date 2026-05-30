package codegen

import "github.com/quant5-lab/runner/ast"

type BuiltinUsageDetector struct {
	targetNames       map[string]bool
	targetMemberExprs map[string]bool
}

func NewBuiltinUsageDetector(names []string) *BuiltinUsageDetector {
	targets := make(map[string]bool, len(names))
	for _, n := range names {
		targets[n] = true
	}
	return &BuiltinUsageDetector{targetNames: targets}
}

/* memberKeys use "obj.prop" format (e.g., "session.isfirstbar") */
func NewBuiltinUsageDetectorWithMembers(names []string, memberKeys []string) *BuiltinUsageDetector {
	targets := make(map[string]bool, len(names))
	for _, n := range names {
		targets[n] = true
	}
	members := make(map[string]bool, len(memberKeys))
	for _, m := range memberKeys {
		members[m] = true
	}
	return &BuiltinUsageDetector{targetNames: targets, targetMemberExprs: members}
}

func (d *BuiltinUsageDetector) Detect(program *ast.Program) map[string]bool {
	if program == nil {
		return nil
	}

	found := make(map[string]bool)
	for _, node := range program.Body {
		d.scanNode(node, found)
	}
	return found
}

func (d *BuiltinUsageDetector) scanNode(node ast.Node, found map[string]bool) {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			d.scanExpression(decl.Init, found)
		}
	case *ast.ExpressionStatement:
		d.scanExpression(n.Expression, found)
	case *ast.IfStatement:
		d.scanExpression(n.Test, found)
		for _, stmt := range n.Consequent {
			d.scanNode(stmt, found)
		}
		for _, stmt := range n.Alternate {
			d.scanNode(stmt, found)
		}
	case *ast.ForStatement:
		d.scanExpression(n.From, found)
		d.scanExpression(n.To, found)
		d.scanExpression(n.Step, found)
		for _, stmt := range n.Body {
			d.scanNode(stmt, found)
		}
	case *ast.ForInStatement:
		d.scanExpression(n.Collection, found)
		for _, stmt := range n.Body {
			d.scanNode(stmt, found)
		}
	case *ast.WhileStatement:
		d.scanExpression(n.Condition, found)
		for _, stmt := range n.Body {
			d.scanNode(stmt, found)
		}
	}
}

func (d *BuiltinUsageDetector) scanExpression(expr ast.Expression, found map[string]bool) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		if d.targetNames[e.Name] {
			found[e.Name] = true
		}
	case *ast.MemberExpression:
		if ident, ok := e.Object.(*ast.Identifier); ok {
			if d.targetNames[ident.Name] {
				found[ident.Name] = true
			}
			if prop, propOk := e.Property.(*ast.Identifier); propOk && d.targetMemberExprs != nil {
				key := ident.Name + "." + prop.Name
				if d.targetMemberExprs[key] {
					found[key] = true
				}
			}
		}
		/* Handle subscript of member expression: session.isfirstbar[1] */
		if innerMember, ok := e.Object.(*ast.MemberExpression); ok && e.Computed {
			if innerObj, objOk := innerMember.Object.(*ast.Identifier); objOk {
				if innerProp, propOk := innerMember.Property.(*ast.Identifier); propOk && d.targetMemberExprs != nil {
					key := innerObj.Name + "." + innerProp.Name
					if d.targetMemberExprs[key] {
						found[key] = true
					}
				}
			}
		}
		d.scanExpression(e.Object, found)
		d.scanExpression(e.Property, found)
	case *ast.CallExpression:
		d.scanExpression(e.Callee, found)
		for _, arg := range e.Arguments {
			d.scanExpression(arg, found)
		}
	case *ast.BinaryExpression:
		d.scanExpression(e.Left, found)
		d.scanExpression(e.Right, found)
	case *ast.LogicalExpression:
		d.scanExpression(e.Left, found)
		d.scanExpression(e.Right, found)
	case *ast.ConditionalExpression:
		d.scanExpression(e.Test, found)
		d.scanExpression(e.Consequent, found)
		d.scanExpression(e.Alternate, found)
	case *ast.UnaryExpression:
		d.scanExpression(e.Argument, found)
	case *ast.IfStatement:
		// Pine `if` used as rvalue: VariableDeclarator.Init can be *ast.IfStatement.
		d.scanExpression(e.Test, found)
		for _, stmt := range e.Consequent {
			d.scanNode(stmt, found)
		}
		for _, stmt := range e.Alternate {
			d.scanNode(stmt, found)
		}
	case *ast.ArrowFunctionExpression:
		for _, bodyNode := range e.Body {
			d.scanNode(bodyNode, found)
			if expr, ok := bodyNode.(ast.Expression); ok {
				d.scanExpression(expr, found)
			}
		}
	}
}
