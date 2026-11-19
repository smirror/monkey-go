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
		copyNode.Statements = make([]Statement, len(n.Statements))
		for i, statement := range n.Statements {
			if statement != nil {
				copyNode.Statements[i] = Modify(statement, fn).(Statement)
			}
		}
		node = &copyNode
	case *ExpressionStatement:
		copyNode := *n
		if n.Expression != nil {
			copyNode.Expression = Modify(n.Expression, fn).(Expression)
		}
		node = &copyNode
	case *LetStatement:
		copyNode := *n
		if n.Value != nil {
			copyNode.Value = Modify(n.Value, fn).(Expression)
		}
		node = &copyNode
	case *ReturnStatement:
		copyNode := *n
		if n.ReturnValue != nil {
			copyNode.ReturnValue = Modify(n.ReturnValue, fn).(Expression)
		}
		node = &copyNode
	case *ForInStatement:
		copyNode := *n
		if n.Iterable != nil {
			copyNode.Iterable = Modify(n.Iterable, fn).(Expression)
		}
		if n.Body != nil {
			copyNode.Body = Modify(n.Body, fn).(*BlockStatement)
		}
		node = &copyNode
	case *BlockStatement:
		copyNode := *n
		copyNode.Statements = make([]Statement, len(n.Statements))
		for i, statement := range n.Statements {
			if statement != nil {
				copyNode.Statements[i] = Modify(statement, fn).(Statement)
			}
		}
		node = &copyNode
	case *PrefixExpression:
		copyNode := *n
		if n.Right != nil {
			copyNode.Right = Modify(n.Right, fn).(Expression)
		}
		node = &copyNode
	case *InfixExpression:
		copyNode := *n
		if n.Left != nil {
			copyNode.Left = Modify(n.Left, fn).(Expression)
		}
		if n.Right != nil {
			copyNode.Right = Modify(n.Right, fn).(Expression)
		}
		node = &copyNode
	case *IfExpression:
		copyNode := *n
		if n.Condition != nil {
			copyNode.Condition = Modify(n.Condition, fn).(Expression)
		}
		if n.Consequence != nil {
			copyNode.Consequence = Modify(n.Consequence, fn).(*BlockStatement)
		}
		if n.Alternative != nil {
			copyNode.Alternative = Modify(n.Alternative, fn).(*BlockStatement)
		}
		node = &copyNode
	case *FunctionLiteral:
		copyNode := *n
		if n.Body != nil {
			copyNode.Body = Modify(n.Body, fn).(*BlockStatement)
		}
		node = &copyNode
	case *MacroLiteral:
		copyNode := *n
		if n.Body != nil {
			copyNode.Body = Modify(n.Body, fn).(*BlockStatement)
		}
		node = &copyNode
	case *CallExpression:
		copyNode := *n
		if n.Function != nil {
			copyNode.Function = Modify(n.Function, fn).(Expression)
		}
		copyNode.Arguments = make([]Expression, len(n.Arguments))
		for i, arg := range n.Arguments {
			if arg != nil {
				copyNode.Arguments[i] = Modify(arg, fn).(Expression)
			}
		}
		node = &copyNode
	case *ArrayLiteral:
		copyNode := *n
		copyNode.Elements = make([]Expression, len(n.Elements))
		for i, element := range n.Elements {
			if element != nil {
				copyNode.Elements[i] = Modify(element, fn).(Expression)
			}
		}
		node = &copyNode
	case *IndexExpression:
		copyNode := *n
		if n.Left != nil {
			copyNode.Left = Modify(n.Left, fn).(Expression)
		}
		if n.Index != nil {
			copyNode.Index = Modify(n.Index, fn).(Expression)
		}
		node = &copyNode
	case *HashMapLiteral:
		copyNode := *n
		copyNode.Pairs = make(map[Expression]Expression, len(n.Pairs))
		for key, value := range n.Pairs {
			newKey := Modify(key, fn).(Expression)
			newValue := Modify(value, fn).(Expression)
			copyNode.Pairs[newKey] = newValue
		}
		node = &copyNode
	}
	return fn(node)
}
