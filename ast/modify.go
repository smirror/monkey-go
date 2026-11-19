package ast

// Modifier transforms AST nodes. Modify traverses children before their parents
// and copies composite nodes so reusable macro definitions are not mutated.
type Modifier func(Node) Node

func Modify(node Node, fn Modifier) Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *Program:
		copyNode := *n
		copyNode.Statements = modifyNodes(n.Statements, fn)
		node = &copyNode
	case Statement:
		node = modifyStatementChildren(n, fn)
	case Expression:
		node = modifyExpressionChildren(n, fn)
	}
	return fn(node)
}

func modifyStatementChildren(node Statement, fn Modifier) Statement {
	switch n := node.(type) {
	case *ExpressionStatement:
		copyNode := *n
		copyNode.Expression = modifyExpression(n.Expression, fn)
		return &copyNode
	case *LetStatement:
		copyNode := *n
		copyNode.Value = modifyExpression(n.Value, fn)
		return &copyNode
	case *ReturnStatement:
		copyNode := *n
		copyNode.ReturnValue = modifyExpression(n.ReturnValue, fn)
		return &copyNode
	case *ForInStatement:
		copyNode := *n
		copyNode.Iterable = modifyExpression(n.Iterable, fn)
		copyNode.Body = modifyBlock(n.Body, fn)
		return &copyNode
	case *BlockStatement:
		copyNode := *n
		copyNode.Statements = modifyNodes(n.Statements, fn)
		return &copyNode
	}
	return node
}

func modifyExpressionChildren(node Expression, fn Modifier) Expression {
	switch n := node.(type) {
	case *PrefixExpression:
		copyNode := *n
		copyNode.Right = modifyExpression(n.Right, fn)
		return &copyNode
	case *InfixExpression:
		copyNode := *n
		copyNode.Left = modifyExpression(n.Left, fn)
		copyNode.Right = modifyExpression(n.Right, fn)
		return &copyNode
	case *IfExpression:
		copyNode := *n
		copyNode.Condition = modifyExpression(n.Condition, fn)
		copyNode.Consequence = modifyBlock(n.Consequence, fn)
		copyNode.Alternative = modifyBlock(n.Alternative, fn)
		return &copyNode
	case *CallExpression:
		copyNode := *n
		copyNode.Function = modifyExpression(n.Function, fn)
		copyNode.Arguments = modifyNodes(n.Arguments, fn)
		return &copyNode
	case *IndexExpression:
		copyNode := *n
		copyNode.Left = modifyExpression(n.Left, fn)
		copyNode.Index = modifyExpression(n.Index, fn)
		return &copyNode
	}
	return modifyLiteralChildren(node, fn)
}

func modifyLiteralChildren(node Expression, fn Modifier) Expression {
	switch n := node.(type) {
	case *FunctionLiteral:
		copyNode := *n
		copyNode.Body = modifyBlock(n.Body, fn)
		return &copyNode
	case *MacroLiteral:
		copyNode := *n
		copyNode.Body = modifyBlock(n.Body, fn)
		return &copyNode
	case *ArrayLiteral:
		copyNode := *n
		copyNode.Elements = modifyNodes(n.Elements, fn)
		return &copyNode
	case *HashMapLiteral:
		copyNode := *n
		copyNode.Pairs = make(map[Expression]Expression, len(n.Pairs))
		for key, value := range n.Pairs {
			newKey := Modify(key, fn).(Expression)
			copyNode.Pairs[newKey] = Modify(value, fn).(Expression)
		}
		return &copyNode
	}
	return node
}

func modifyNodes[T Node](nodes []T, fn Modifier) []T {
	rewritten := make([]T, len(nodes))
	for i, node := range nodes {
		if Node(node) != nil {
			rewritten[i] = Modify(node, fn).(T)
		}
	}
	return rewritten
}

func modifyExpression(node Expression, fn Modifier) Expression {
	if node == nil {
		return nil
	}
	return Modify(node, fn).(Expression)
}

func modifyBlock(node *BlockStatement, fn Modifier) *BlockStatement {
	if node == nil {
		return nil
	}
	return Modify(node, fn).(*BlockStatement)
}
