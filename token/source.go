package token

// Position identifies a boundary in a source file. Offset counts UTF-8 bytes;
// Line and Column are one-based, and Column counts runes (a tab counts as one).
type Position struct {
	Offset int
	Line   int
	Column int
}

// Source is a snapshot shared by tokens and AST nodes. Treat it as immutable:
// retained functions must keep the text of the submission that defined them.
type Source struct {
	Name string
	Text string
}

// Span is a half-open range [Start, End) in Source. Its zero value means that
// no source location is available, for example for a manually constructed AST.
type Span struct {
	Source *Source
	Start  Position
	End    Position
}
