package ast

import "monkey-go/token"

// SourceLocation preserves the complete parsed range without changing a node's
// token, which still identifies its keyword, literal, or operator precisely.
type SourceLocation struct {
	span token.Span
}

func (l SourceLocation) Span() token.Span {
	return l.span
}

func (l *SourceLocation) setSpan(span token.Span) {
	l.span = span
}

// SetSpan attaches a parsed range, including grouping delimiters, to a node.
// A nil node is ignored so failed parses do not need a separate location path.
func SetSpan(node Node, span token.Span) {
	if located, ok := node.(interface{ setSpan(token.Span) }); ok {
		located.setSpan(span)
	}
}
