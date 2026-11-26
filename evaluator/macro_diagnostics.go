package evaluator

import (
	"fmt"
	"monkey-go/diagnostic"
	"monkey-go/token"
)

// MacroExpansionError preserves source information while satisfying the error
// return type of ExpandMacros. Error retains the existing message-only API.
type MacroExpansionError struct {
	report diagnostic.Diagnostic
}

func (e *MacroExpansionError) Error() string {
	return e.report.Message
}

func (e *MacroExpansionError) Diagnostic() diagnostic.Diagnostic {
	report := e.report
	report.Stack = append([]diagnostic.Frame(nil), report.Stack...)
	return report
}

func newMacroExpansionError(span token.Span, format string, args ...any) *MacroExpansionError {
	return &MacroExpansionError{report: diagnostic.Diagnostic{
		Kind: diagnostic.Macro, Message: fmt.Sprintf(format, args...), Span: span,
	}}
}
