package evaluator

import "monkey-go/ast"

type macroScope map[string]bool
type scopedMacroRewrite func(ast.Node, macroScope) (ast.Node, bool)

// macroSyntaxWalker keeps binder names out of template substitution and macro
// lookup. Each function and loop gets its own bindings; lets shadow references
// after their initializers in the surrounding sequence of statements.
type macroSyntaxWalker struct {
	scope   macroScope
	rewrite scopedMacroRewrite
}

func rewriteMacroSyntax(node ast.Node, rewrite scopedMacroRewrite) ast.Node {
	walker := macroSyntaxWalker{scope: make(macroScope), rewrite: rewrite}
	return walker.walk(node)
}

func (w macroSyntaxWalker) walk(node ast.Node) ast.Node {
	return ast.Rewrite(node, func(current ast.Node) (ast.Node, bool) {
		rewritten, descend := w.rewrite(current, w.scope)
		if !descend {
			return rewritten, false
		}
		return w.walkBindings(rewritten)
	})
}

func (w macroSyntaxWalker) walkBindings(node ast.Node) (ast.Node, bool) {
	switch node := node.(type) {
	case *ast.Program:
		copyNode := *node
		copyNode.Statements = w.statements(node.Statements)
		return &copyNode, false
	case *ast.BlockStatement:
		return w.block(node), false
	case *ast.LetStatement:
		return w.letStatement(node), false
	case *ast.FunctionLiteral:
		copyNode := *node
		copyNode.Parameters, copyNode.Body = w.functionBody(node.Parameters, node.Body)
		return &copyNode, false
	case *ast.MacroLiteral:
		copyNode := *node
		copyNode.Parameters, copyNode.Body = w.functionBody(node.Parameters, node.Body)
		return &copyNode, false
	case *ast.ForInStatement:
		return w.forStatement(node), false
	case *ast.IfExpression:
		return w.ifExpression(node), false
	}
	return node, true
}

func (w macroSyntaxWalker) statements(statements []ast.Statement) []ast.Statement {
	result := make([]ast.Statement, len(statements))
	for i, statement := range statements {
		if statement != nil {
			result[i] = w.walk(statement).(ast.Statement)
		}
	}
	return result
}

func (w macroSyntaxWalker) block(block *ast.BlockStatement) *ast.BlockStatement {
	if block == nil {
		return nil
	}
	copyNode := *block
	copyNode.Statements = w.statements(block.Statements)
	return &copyNode
}

func (w macroSyntaxWalker) expression(expression ast.Expression) ast.Expression {
	if expression == nil {
		return nil
	}
	return w.walk(expression).(ast.Expression)
}

func (w macroSyntaxWalker) child(names ...*ast.Identifier) macroSyntaxWalker {
	child := macroSyntaxWalker{scope: make(macroScope, len(w.scope)+len(names)), rewrite: w.rewrite}
	for name := range w.scope {
		child.scope[name] = true
	}
	for _, name := range names {
		child.scope[name.Value] = true
	}
	return child
}

func (w macroSyntaxWalker) letStatement(node *ast.LetStatement) *ast.LetStatement {
	copyNode := *node
	copyNode.Name = cloneMacroSyntax(node.Name).(*ast.Identifier)
	initializer := w
	switch node.Value.(type) {
	case *ast.FunctionLiteral, *ast.MacroLiteral:
		// A named function resolves recursive references after it is bound.
		initializer = w.child(node.Name)
	}
	copyNode.Value = initializer.expression(node.Value)
	w.scope[node.Name.Value] = true
	return &copyNode
}

func (w macroSyntaxWalker) functionBody(parameters []*ast.Identifier, body *ast.BlockStatement) ([]*ast.Identifier, *ast.BlockStatement) {
	copied := make([]*ast.Identifier, len(parameters))
	for i, parameter := range parameters {
		copied[i] = cloneMacroSyntax(parameter).(*ast.Identifier)
	}
	return copied, w.child(parameters...).block(body)
}

func (w macroSyntaxWalker) forStatement(node *ast.ForInStatement) *ast.ForInStatement {
	copyNode := *node
	copyNode.Variable = cloneMacroSyntax(node.Variable).(*ast.Identifier)
	copyNode.Iterable = w.expression(node.Iterable)
	copyNode.Body = w.child(node.Variable).block(node.Body)
	return &copyNode
}

func (w macroSyntaxWalker) ifExpression(node *ast.IfExpression) *ast.IfExpression {
	copyNode := *node
	copyNode.Condition = w.expression(node.Condition)
	// Neither branch is executed during expansion. Keep its bindings local to
	// its syntax so an untaken branch cannot hide references in its sibling.
	copyNode.Consequence = w.child().block(node.Consequence)
	copyNode.Alternative = w.child().block(node.Alternative)
	return &copyNode
}
