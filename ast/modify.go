package ast

// Modifier transforms AST nodes, visiting children before their parents.
// It preserves statement/expression node categories.
type Modifier func(Node) Node

func Modify(node Node, fn Modifier) Node {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *Program:
		for i, statement := range n.Statements {
			if statement != nil {
				n.Statements[i] = Modify(statement, fn).(Statement)
			}
		}
	case *ExpressionStatement:
		if n.Expression != nil {
			n.Expression = Modify(n.Expression, fn).(Expression)
		}
	case *LetStatement:
		if n.Value != nil {
			n.Value = Modify(n.Value, fn).(Expression)
		}
	case *ReturnStatement:
		if n.ReturnValue != nil {
			n.ReturnValue = Modify(n.ReturnValue, fn).(Expression)
		}
	case *ForInStatement:
		if n.Iterable != nil {
			n.Iterable = Modify(n.Iterable, fn).(Expression)
		}
		if n.Body != nil {
			n.Body = Modify(n.Body, fn).(*BlockStatement)
		}
	case *BlockStatement:
		for i, statement := range n.Statements {
			if statement != nil {
				n.Statements[i] = Modify(statement, fn).(Statement)
			}
		}
	case *PrefixExpression:
		if n.Right != nil {
			n.Right = Modify(n.Right, fn).(Expression)
		}
	case *InfixExpression:
		if n.Left != nil {
			n.Left = Modify(n.Left, fn).(Expression)
		}
		if n.Right != nil {
			n.Right = Modify(n.Right, fn).(Expression)
		}
	case *IfExpression:
		if n.Condition != nil {
			n.Condition = Modify(n.Condition, fn).(Expression)
		}
		if n.Consequence != nil {
			n.Consequence = Modify(n.Consequence, fn).(*BlockStatement)
		}
		if n.Alternative != nil {
			n.Alternative = Modify(n.Alternative, fn).(*BlockStatement)
		}
	case *FunctionLiteral:
		if n.Body != nil {
			n.Body = Modify(n.Body, fn).(*BlockStatement)
		}
	case *MacroLiteral:
		if n.Body != nil {
			n.Body = Modify(n.Body, fn).(*BlockStatement)
		}
	case *CallExpression:
		if n.Function != nil {
			n.Function = Modify(n.Function, fn).(Expression)
		}
		for i, arg := range n.Arguments {
			if arg != nil {
				n.Arguments[i] = Modify(arg, fn).(Expression)
			}
		}
	case *ArrayLiteral:
		for i, element := range n.Elements {
			if element != nil {
				n.Elements[i] = Modify(element, fn).(Expression)
			}
		}
	case *IndexExpression:
		if n.Left != nil {
			n.Left = Modify(n.Left, fn).(Expression)
		}
		if n.Index != nil {
			n.Index = Modify(n.Index, fn).(Expression)
		}
	case *HashMapLiteral:
		pairs := make(map[Expression]Expression, len(n.Pairs))
		for key, value := range n.Pairs {
			newKey := Modify(key, fn).(Expression)
			newValue := Modify(value, fn).(Expression)
			pairs[newKey] = newValue
		}
		n.Pairs = pairs
	}
	return fn(node)
}
