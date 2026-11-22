package diagnostic_test

import (
	"errors"
	"monkey-go/diagnostic"
	"monkey-go/token"
	"strings"
	"testing"
)

type sourceFormatCase struct {
	name     string
	text     string
	start    token.Position
	end      token.Position
	location string
	line     string
	marker   string
}

func TestFormatSourcePositionsAndDisplayWidths(t *testing.T) {
	checkSourceFormats(t, []sourceFormatCase{
		{
			name: "ASCII expression", text: "1 + missing",
			start: position(4, 1, 5), end: position(11, 1, 12),
			location: "1:5", line: "1 + missing", marker: "    ^~~~~~~",
		},
		{
			name: "wide characters before and within span", text: "日 + 本",
			start: position(6, 1, 5), end: position(9, 1, 6),
			location: "1:5", line: "日 + 本", marker: "     ^~",
		},
		{
			name: "combining mark before span", text: "e\u0301 + x",
			start: position(6, 1, 6), end: position(7, 1, 7),
			location: "1:6", line: "e\u0301 + x", marker: "    ^",
		},
		{
			name: "combining sequence within span", text: "e\u0301",
			start: position(0, 1, 1), end: position(3, 1, 3),
			location: "1:1", line: "e\u0301", marker: "^",
		},
		{
			name: "tab is four spaces even after text", text: "a\tx",
			start: position(2, 1, 3), end: position(3, 1, 4),
			location: "1:3", line: "a    x", marker: "     ^",
		},
	})
}

func TestFormatSourceLineEndings(t *testing.T) {
	checkSourceFormats(t, []sourceFormatCase{
		{
			name: "LF line ending", text: "one\ntwo",
			start: position(4, 2, 1), end: position(7, 2, 4),
			location: "2:1", line: "two", marker: "^~~",
		},
		{
			name: "CRLF line ending", text: "one\r\ntwo",
			start: position(5, 2, 1), end: position(8, 2, 4),
			location: "2:1", line: "two", marker: "^~~",
		},
		{
			name: "CR line ending", text: "one\rtwo",
			start: position(4, 2, 1), end: position(7, 2, 4),
			location: "2:1", line: "two", marker: "^~~",
		},
	})
}

func TestFormatEOFAndMultilineSpans(t *testing.T) {
	checkSourceFormats(t, []sourceFormatCase{
		{
			name: "EOF after text", text: "abc",
			start: position(3, 1, 4), end: position(3, 1, 4),
			location: "1:4", line: "abc", marker: "   ^",
		},
		{
			name: "EOF in empty source", text: "",
			start: position(0, 1, 1), end: position(0, 1, 1),
			location: "1:1", line: "", marker: "^",
		},
		{
			name: "EOF after LF", text: "abc\n",
			start: position(4, 2, 1), end: position(4, 2, 1),
			location: "2:1", line: "", marker: "^",
		},
		{
			name: "EOF after CRLF", text: "abc\r\n",
			start: position(5, 2, 1), end: position(5, 2, 1),
			location: "2:1", line: "", marker: "^",
		},
		{
			name: "multiline span ends at first line", text: "one two\nthree",
			start: position(4, 1, 5), end: position(13, 2, 6),
			location: "1:5", line: "one two", marker: "    ^~~",
		},
	})
}

func checkSourceFormats(t *testing.T, tests []sourceFormatCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			d := diagnostic.Diagnostic{
				Kind: diagnostic.Runtime, Message: "boom",
				Span: token.Span{
					Source: &token.Source{Name: "example.mk", Text: tt.text},
					Start:  tt.start, End: tt.end,
				},
			}
			want := "example.mk:" + tt.location + ": runtime error: boom\n" +
				"  | " + tt.line + "\n  | " + tt.marker + "\n"
			if got := diagnostic.Format(d); got != want {
				t.Fatalf("formatted diagnostic:\n got %q\nwant %q", got, want)
			}
		})
	}
}

func TestFormatWithoutSource(t *testing.T) {
	tests := []struct {
		name string
		kind diagnostic.Kind
		span token.Span
		want string
	}{
		{name: "no kind or position", want: "error: boom\n"},
		{name: "syntax kind", kind: diagnostic.Syntax, want: "syntax error: boom\n"},
		{name: "runtime kind", kind: diagnostic.Runtime, want: "runtime error: boom\n"},
		{name: "input kind", kind: diagnostic.IO, want: "io error: boom\n"},
		{
			name: "position without source", kind: diagnostic.Runtime,
			span: token.Span{Start: position(0, 2, 3)},
			want: "<input>:2:3: runtime error: boom\n",
		},
		{
			name: "source without name", kind: diagnostic.Runtime,
			span: token.Span{
				Source: &token.Source{Text: "x"},
				Start:  position(0, 1, 1), End: position(1, 1, 2),
			},
			want: "<input>:1:1: runtime error: boom\n  | x\n  | ^\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diagnostic.Format(diagnostic.Diagnostic{Kind: tt.kind, Message: "boom", Span: tt.span})
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatOmitsExcerptForInvalidPositions(t *testing.T) {
	tests := []struct {
		name  string
		start token.Position
		end   token.Position
	}{
		{name: "negative offset", start: position(-1, 1, 1), end: position(0, 1, 1)},
		{name: "offset past EOF", start: position(5, 1, 1), end: position(5, 1, 1)},
		{name: "zero line", start: position(0, 0, 1), end: position(3, 1, 2)},
		{name: "zero column", start: position(0, 1, 0), end: position(3, 1, 2)},
		{name: "start splits UTF-8 rune", start: position(1, 1, 1), end: position(3, 1, 2)},
		{name: "end splits UTF-8 rune", start: position(0, 1, 1), end: position(2, 1, 2)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := diagnostic.Diagnostic{
				Kind: diagnostic.Runtime, Message: "boom",
				Span: token.Span{
					Source: &token.Source{Name: "example.mk", Text: "日x"},
					Start:  tt.start, End: tt.end,
				},
			}
			got := diagnostic.Format(d)
			if !strings.Contains(got, "runtime error: boom\n") {
				t.Fatalf("message was lost: %q", got)
			}
			if strings.Contains(got, "  | ") {
				t.Fatalf("invalid position produced a misleading excerpt: %q", got)
			}
		})
	}
}

func TestFormatClipsInconsistentEndOffsets(t *testing.T) {
	tests := []struct {
		name   string
		end    token.Position
		marker string
	}{
		{name: "end before start", end: position(0, 1, 1), marker: " ^"},
		{name: "end past EOF", end: position(100, 1, 101), marker: " ^~"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := diagnostic.Diagnostic{
				Kind: diagnostic.Runtime, Message: "boom",
				Span: token.Span{
					Source: &token.Source{Name: "example.mk", Text: "abc"},
					Start:  position(1, 1, 2), End: tt.end,
				},
			}
			want := "example.mk:1:2: runtime error: boom\n  | abc\n  | " + tt.marker + "\n"
			if got := diagnostic.Format(d); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}

func TestFormatPreservesInnermostToOutermostStackOrder(t *testing.T) {
	d := diagnostic.Diagnostic{
		Kind: diagnostic.Runtime, Message: "division by zero",
		Span: token.Span{
			Source: &token.Source{Name: "inner.mk", Text: "1 / 0"},
			Start:  position(0, 1, 1), End: position(5, 1, 6),
		},
		Stack: []diagnostic.Frame{
			{
				Name: "inner",
				Span: token.Span{Source: &token.Source{Name: "outer.mk"}, Start: position(0, 2, 3)},
			},
			{
				Name: "outer",
				Span: token.Span{Source: &token.Source{Name: "main.mk"}, Start: position(0, 9, 1)},
			},
			{Name: "host callback"},
		},
	}
	want := "inner.mk:1:1: runtime error: division by zero\n" +
		"  | 1 / 0\n  | ^~~~~\n" +
		"  at inner (outer.mk:2:3)\n" +
		"  at outer (main.mk:9:1)\n" +
		"  at host callback\n"
	if got := diagnostic.Format(d); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatEscapesControlCharactersAndAlignsCaret(t *testing.T) {
	d := diagnostic.Diagnostic{
		Kind: diagnostic.Runtime, Message: "bad\x00\ninput",
		Span: token.Span{
			Source: &token.Source{Name: "bad\n.mk", Text: "a\x1bb"},
			Start:  position(2, 1, 3), End: position(3, 1, 4),
		},
		Stack: []diagnostic.Frame{{Name: "f\r\n\x1b"}},
	}
	want := "bad\\n.mk:1:3: runtime error: bad\\x00\\ninput\n" +
		"  | a\\x1bb\n  |      ^\n" +
		"  at f\\r\\n\\x1b\n"
	if got := diagnostic.Format(d); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWritePropagatesOutputFailure(t *testing.T) {
	wantErr := errors.New("output failed")
	writer := failingWriter{err: wantErr}
	got := diagnostic.Write(writer, diagnostic.Diagnostic{Kind: diagnostic.IO, Message: "read failed"})
	if !errors.Is(got, wantErr) {
		t.Fatalf("got error %v, want %v", got, wantErr)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 3, w.err // Even a failure after partial output must be returned.
}

func position(offset, line, column int) token.Position {
	return token.Position{Offset: offset, Line: line, Column: column}
}
