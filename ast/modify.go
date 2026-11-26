package ast

// Modifier transforms AST nodes. Modify traverses children before their parents
// and copies nodes so reusable macro definitions are not mutated.
type Modifier func(Node) Node

func Modify(node Node, fn Modifier) Node {
	if node == nil {
		return nil
	}
	return fn(copyChildren(node, func(child Node) Node {
		return Modify(child, fn)
	}))
}

// Rewriter replaces a node before visiting its children. Returning false stops
// traversal at the replacement, which is useful when inserting argument syntax.
// The callback must treat its input as immutable and owns skipped replacements.
type Rewriter func(Node) (Node, bool)

func Rewrite(node Node, fn Rewriter) Node {
	if node == nil {
		return nil
	}
	replacement, descend := fn(node)
	if !descend {
		return replacement
	}
	return copyChildren(replacement, func(child Node) Node {
		return Rewrite(child, fn)
	})
}

func copyChildren(node Node, visit Modifier) Node {
	switch n := node.(type) {
	case *Program:
		copyNode := *n
		copyNode.Statements = modifyNodes(n.Statements, visit)
		return &copyNode
	case Statement:
		return modifyStatementChildren(n, visit)
	case Expression:
		return modifyExpressionChildren(n, visit)
	}
	return node
}

func modifyStatementChildren(node Statement, visit Modifier) Statement {
	switch n := node.(type) {
	case *ExpressionStatement:
		copyNode := *n
		copyNode.Expression = modifyExpression(n.Expression, visit)
		return &copyNode
	case *LetStatement:
		copyNode := *n
		copyNode.Name = copyValue(n.Name)
		copyNode.Value = modifyExpression(n.Value, visit)
		return &copyNode
	case *ReturnStatement:
		copyNode := *n
		copyNode.ReturnValue = modifyExpression(n.ReturnValue, visit)
		return &copyNode
	case *ForInStatement:
		copyNode := *n
		copyNode.Variable = copyValue(n.Variable)
		copyNode.Iterable = modifyExpression(n.Iterable, visit)
		copyNode.Body = modifyBlock(n.Body, visit)
		return &copyNode
	case *BlockStatement:
		copyNode := *n
		copyNode.Statements = modifyNodes(n.Statements, visit)
		return &copyNode
	}
	return node
}

func modifyExpressionChildren(node Expression, visit Modifier) Expression {
	switch n := node.(type) {
	case *PrefixExpression:
		copyNode := *n
		copyNode.Right = modifyExpression(n.Right, visit)
		return &copyNode
	case *InfixExpression:
		copyNode := *n
		copyNode.Left = modifyExpression(n.Left, visit)
		copyNode.Right = modifyExpression(n.Right, visit)
		return &copyNode
	case *IfExpression:
		copyNode := *n
		copyNode.Condition = modifyExpression(n.Condition, visit)
		copyNode.Consequence = modifyBlock(n.Consequence, visit)
		copyNode.Alternative = modifyBlock(n.Alternative, visit)
		return &copyNode
	case *CallExpression:
		copyNode := *n
		copyNode.Function = modifyExpression(n.Function, visit)
		copyNode.Arguments = modifyNodes(n.Arguments, visit)
		return &copyNode
	case *IndexExpression:
		copyNode := *n
		copyNode.Left = modifyExpression(n.Left, visit)
		copyNode.Index = modifyExpression(n.Index, visit)
		return &copyNode
	case *MacroExpansion:
		return modifyExpansion(n, visit)
	}
	return modifyLiteralChildren(node, visit)
}

func modifyLiteralChildren(node Expression, visit Modifier) Expression {
	switch n := node.(type) {
	case *FunctionLiteral:
		copyNode := *n
		copyNode.Parameters = copyIdentifiers(n.Parameters)
		copyNode.Body = modifyBlock(n.Body, visit)
		return &copyNode
	case *MacroLiteral:
		copyNode := *n
		copyNode.Parameters = copyIdentifiers(n.Parameters)
		copyNode.Body = modifyBlock(n.Body, visit)
		return &copyNode
	case *ArrayLiteral:
		copyNode := *n
		copyNode.Elements = modifyNodes(n.Elements, visit)
		return &copyNode
	case *HashMapLiteral:
		copyNode := *n
		copyNode.Pairs = make(map[Expression]Expression, len(n.Pairs))
		for key, value := range n.Pairs {
			newKey := visit(key).(Expression)
			copyNode.Pairs[newKey] = visit(value).(Expression)
		}
		return &copyNode
	}
	return copyLiteral(node)
}

func copyLiteral(node Expression) Expression {
	switch n := node.(type) {
	case *Identifier:
		return copyValue(n)
	case *IntegerLiteral:
		return copyValue(n)
	case *Boolean:
		return copyValue(n)
	case *StringLiteral:
		return copyValue(n)
	}
	return node
}

func modifyExpansion(node *MacroExpansion, visit Modifier) *MacroExpansion {
	copyNode := *node
	copyNode.Body = modifyBlock(node.Body, visit)
	if node.Call != nil {
		copyNode.Call = copyTree(node.Call).(*CallExpression)
	}
	return &copyNode
}

func copyTree(node Node) Node {
	return copyChildren(node, copyTree)
}

func copyValue[T any](node *T) *T {
	if node == nil {
		return nil
	}
	copyNode := *node
	return &copyNode
}

func copyIdentifiers(nodes []*Identifier) []*Identifier {
	if nodes == nil {
		return nil
	}
	copied := make([]*Identifier, len(nodes))
	for i, node := range nodes {
		copied[i] = copyValue(node)
	}
	return copied
}

func modifyNodes[T Node](nodes []T, visit Modifier) []T {
	rewritten := make([]T, len(nodes))
	for i, node := range nodes {
		if Node(node) != nil {
			rewritten[i] = visit(node).(T)
		}
	}
	return rewritten
}

func modifyExpression(node Expression, visit Modifier) Expression {
	if node == nil {
		return nil
	}
	return visit(node).(Expression)
}

func modifyBlock(node *BlockStatement, visit Modifier) *BlockStatement {
	if node == nil {
		return nil
	}
	return visit(node).(*BlockStatement)
}
