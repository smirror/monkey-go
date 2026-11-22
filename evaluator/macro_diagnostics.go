package evaluator

import (
	"fmt"
	"monkey-go/ast"
	"monkey-go/diagnostic"
	"monkey-go/object"
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

func macroEvaluationError(name string, call *ast.CallExpression, cause *object.Error) *MacroExpansionError {
	located := withErrorSpan(cause, call.Span())
	traced := withCallFrame(located, call).(*object.Error)
	report := traced.Diagnostic()
	report.Kind = diagnostic.Macro
	report.Message = fmt.Sprintf("macro %s: %s", name, cause.Message)
	return &MacroExpansionError{report: report}
}
