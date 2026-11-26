# monkey-go
[![codecov](https://codecov.io/gh/smirror/monkey-go/branch/main/graph/badge.svg?token=NQRJCUX1MK)](https://codecov.io/gh/smirror/monkey-go)
[![CodeFactor](https://www.codefactor.io/repository/github/smirror/monkey-go/badge)](https://www.codefactor.io/repository/github/smirror/monkey-go)
[![CircleCI](https://dl.circleci.com/status-badge/img/gh/smirror/monkey-go/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/smirror/monkey-go/tree/main)
[![CodeScene Code Health](https://codescene.io/projects/28045/status-badges/code-health)](https://codescene.io/projects/28045)
[![Actions Status](https://github.com/smirror/monkey-go/workflows/lint/badge.svg)](https://github.com/smirror/monkey-go/actions)

monkey-go(語) is meaning about "Monkey Language".
monkey-go is based on "
[Writing An Interpreter In Go](https://interpreterbook.com/)
".

## Environment
 - go version >= 1.18

## Usage

### Start REPL

```sh
go run main.go
```

```
Hello <user>! This is the Monkey programming language!
Feel free to type in commands
>>
```

Type `exit` to quit. Multi-line `for`, `if`, and function bodies are supported;
the REPL uses a `.. ` continuation prompt until the delimiters close.

### Run tests

```sh
go test ./...
```

## Syntax

### Variables

```monkey
let x = 5;
let name = "Monkey";
let flag = true;
```

### Integers

```monkey
let a = 10;
let b = 3;
a + b  // 13
a - b  // 7
a * b  // 30
a / b  // 3
a % b  // 1
a < b  // false
a <= b // false
a > b  // true
a >= b // true
a == b // false
a != b // true
```

Integer arithmetic uses signed 64-bit values. Division by zero and overflow
(in addition, subtraction, multiplication, division, or unary minus) return
a Monkey runtime error rather than panicking or silently wrapping. Division
truncates toward zero.

`%` computes an integer remainder using the same truncation rule as division.
A nonzero remainder has the left operand's sign: `-5 % 3` is `-2`, while
`5 % -3` is `2`. A zero divisor returns `modulo by zero`. The remainder of
`MinInt64 % -1` is safely `0`, even though `MinInt64 / -1` returns an overflow
error. These remainder rules follow the [Go integer operator specification](https://go.dev/ref/spec#Integer_operators).

`<`, `<=`, `>`, `>=`, and `%` require integer operands; strings, booleans, and
other types are rejected. Comparisons return a Boolean. Comparisons are binary:
write `a <= b && b <= c` to check a range, rather than chaining `a <= b <= c`.

### Operator Precedence

Binary operators are listed from highest to lowest precedence. Operators on
the same row associate left to right, except assignment, which associates
right to left. Function calls, indexing, and unary operators bind more tightly
than these binary operators; parentheses can override the grouping.

| Operators | Meaning |
|-----------|---------|
| `*`, `/`, `%` | Multiplication, division, remainder |
| `+`, `-` | Addition, subtraction |
| `<`, `<=`, `>`, `>=` | Ordered comparisons |
| `==`, `!=` | Equality comparisons |
| `&&` | Logical AND |
| `\|\|` | Logical OR |
| `=` | Assignment |

For example, `20 % 6 * 3` evaluates to `6`, and `1 + 2 <= 7 % 4` evaluates
to `true`.

### Booleans

```monkey
true
false
!true  // false
true == false  // false
true != false  // true
true && false  // false
true || false  // true
```

`&&` and `||` use short-circuit evaluation: the right-hand expression is
only evaluated when needed. `&&` has higher precedence than `||`, and both
operators return a Boolean (`true` or `false`), not the original operand.

### Strings

Strings support Unicode (including multibyte characters).

```monkey
let s = "Hello, World!";
let greeting = "こんにちは";
s + " Goodbye!"  // concatenation
```

### If / Else

```monkey
if (x > 5) {
    x
} else {
    0
}
```

`if` is an expression and returns a value.

### For Loops

Monkey supports one form of for loop: `for name in iterable { ... }`.
The body must be enclosed in `{}`. The iterable can be a range, an array,
or an expression that evaluates to one of them.

```monkey
for i in range(2, 10) {
    print(i);
}
// prints 2, 3, ... 9

for value in [10, 20, 30] {
    print(value);
}
```

`range(end)`, `range(start, end)`, and `range(start, end, step)` take
integer arguments. The stop value is exclusive; the default step is 1,
and step cannot be zero. Negative steps iterate downward. Ranges are
lazy and do not allocate an array.

The loop variable belongs to a loop-local scope, so an outer variable
with the same name is not overwritten. Assignments to other variables
in outer scopes still work. A loop produces `null`; `return` and
errors propagate immediately. C-style three-clause `for`, `++`,
`break`, `continue`, and `while` are not supported.

Single-line `//` comments are supported; `/* ... */` comments are not.

### Functions

```monkey
let add = fn(x, y) {
    x + y
};

add(3, 4)  // 7
```

Functions are first-class values. The last evaluated expression is the return value.
Function calls require exactly the declared number of arguments; otherwise
they return a Monkey runtime error. An empty function body returns `null`.


```monkey
let factorial = fn(n) {
    if (n == 0) {
        1
    } else {
        n * factorial(n - 1)
    }
};
```

### Macros (experimental)

The body of a `macro` is a syntax template. Before runtime evaluation, each
macro call substitutes its argument expressions for uses of the corresponding
parameters in that template. Calls use the same `name(arguments)` syntax as
ordinary functions.

```monkey
let twice = macro(x) { x + x; };
twice(3 + 1)  // expands to ((3 + 1) + (3 + 1)), yielding 8

let reverse = macro(a, b) { b - a; };
reverse(2 + 2, 10 - 5)  // evaluates (10 - 5) - (2 + 2), yielding 1
```

Arguments are syntax, so their evaluation depends on where the template uses
them. An unused argument is neither expanded nor evaluated. An argument used
twice is evaluated twice when both occurrences run. Runtime control flow still
applies: an expression in an untaken `if` branch is not evaluated.

```monkey
let ignore = macro(unused) { 42; };
ignore(missing)  // 42; missing is never evaluated

let counter = 0;
twice(counter = counter + 1)  // 3; counter is incremented twice and ends at 2
```

Templates can contain multiple statements, `if`, `for`, and `fn`. Their bodies
run in the caller's scope. A `let` in a template creates a binding in that scope,
and a free name is resolved there. A `return` exits the enclosing caller
function; the macro does not introduce a separate function call boundary.

```monkey
let remember = macro(value) { let saved = value; saved + 1; };
remember(40)  // 41
saved         // 40; the binding remains in the caller's scope

let addOffset = macro(value) { value + offset; };
let calculate = fn(offset) { addOffset(2); };
calculate(40)  // 42; offset belongs to calculate's call
```

Macro declarations are currently top-level only. They persist between REPL
inputs and are not runtime function values. Parameter substitution respects
bindings introduced by `let`, function parameters, and loop variables.
Expansion is not hygienic: names introduced by the template can capture or
overwrite caller bindings. Expansion performs syntax substitution only; there
is no expansion-time evaluation of arithmetic, conditions, or helper functions.

Binding analysis follows the syntax being expanded. A declaration within an
`if` branch shadows template parameters and macro names within that branch;
neither branch is executed to decide bindings after the conditional. Runtime
bindings from earlier REPL inputs are also outside this analysis. Use distinct
names for macros and runtime values, and avoid conditionally rebinding template
parameters or macro names.

This is a breaking change from the earlier `quote` / `unquote` macro syntax.
The wrappers must be removed when migrating a template:

```monkey
// Before: obsolete syntax
let twice = macro(x) { quote(unquote(x) + unquote(x)); };

// After
let twice = macro(x) { x + x; };
```

`quote` and `unquote` no longer have special behavior. They are ordinary
identifiers and can be used as names for user-defined functions.

Incorrect argument counts and the 32-level nested expansion limit produce
macro expansion errors. Errors in the expanded body, such as division by zero,
are runtime errors. Both use the REPL diagnostic stream, and the REPL can
accept another submission afterward. Template nodes retain the definition's
source locations, substituted arguments retain their caller's source, and
runtime diagnostics include the macro call site. Expansion does not overwrite
the locations in a reusable definition.

For embedding, `evaluator.ExpandMacros` still returns `(ast.Node, error)`.
An `*evaluator.MacroExpansionError` exposes `Diagnostic()` with kind `macro`,
while its `Error()` method preserves the message-only form.

### Return

```monkey
let earlyReturn = fn(x) {
    if (x > 10) {
        return x;
    }
    x * 2
};
```

### Arrays

```monkey
let a = [1, 2, 3, "four", true];
a[0]   // 1
a[3]   // "four"
```

### Hashes

```monkey
let h = {"key": "value", 1: "one", true: "yes"};
h["key"]  // "value"
h[1]      // "one"
h[true]   // "yes"
```

### Built-in Functions

| Function | Description |
|----------|-------------|
| `len(s)` | Length of string (Unicode-aware) or array |
| `first(arr)` | First element of array |
| `last(arr)` | Last element of array |
| `rest(arr)` | New array without the first element |
| `push(arr, val)` | New array with val appended |
| `print(...)` | Print values to stdout |
| `range(...)` | Lazy numeric range: `range(end)`, `range(start, end)`, or `range(start, end, step)` |

```monkey
let arr = [1, 2, 3];
len(arr)        // 3
first(arr)      // 1
last(arr)       // 3
rest(arr)       // [2, 3]
push(arr, 4)    // [1, 2, 3, 4]

len("hello")    // 5
len("こんにちは")  // 5

print("Hello!", 42, true)
```

## Source Locations

Tokens and parsed AST nodes retain a `token.Span` with a source name, the original
source text, and a half-open range `[Start, End)`. `Offset` is a zero-based UTF-8
byte offset; `Line` and `Column` are one-based, and columns count Unicode runes
(including a tab as one column). CRLF counts as one newline; lone CR and LF also
start a new line. EOF has a stable, zero-width span.

Use `lexer.NewWithFilename(filename, input)` to name a source; `lexer.New(input)`
remains available with an unnamed source. `Parser.Diagnostics()` returns syntax
errors with spans, while `Parser.Errors()` preserves the message-only API.
Unterminated strings and illegal characters are diagnosed instead of being
accepted or silently ending the input.

`ast.Node.Span()` covers the parsed expression or statement, including its
grouping/closing delimiters and a trailing semicolon when present. The node's
`Token.Span` still identifies its own literal, keyword, or operator. Sources are
immutable snapshots shared by nodes, so an earlier REPL definition keeps its
original text when later submissions are parsed. Manually constructed AST nodes
have an unknown (zero) span unless one is attached with `ast.SetSpan`.

## Diagnostics

Parser and runtime errors carry a kind, message, source span, and (for runtime
errors) the entered user-function call sites. `Parser.Errors()` and
`object.Error.Message`/`Inspect()` retain their message-only APIs. Use
`Parser.Diagnostics()` or `object.Error.Diagnostic()` for structured information,
and `diagnostic.Format` / `diagnostic.Write` to render it.

The command-line REPL writes diagnostics to stderr, and prompts/results to
stdout. `repl.Start(in, out)` retains combined output for compatibility;
`repl.StartWithErrorWriter(in, out, errOut)` separates diagnostics. The `print`
builtin still writes to process stdout; routing all builtin IO through a host
writer is tracked separately in #208.

Each complete REPL submission has a source name such as `<repl:1>`. A function
keeps the source that defined it, even when it is called by a later submission:

```text
>> let divide = fn(x) {
..   10 / x
.. };
fn(x){
(10 / x)
}
>> divide(0)
<repl:1>:2:6: runtime error: division by zero
  |   10 / x
  |      ^
  at divide (<repl:2>:1:1)
```

Frames run from the innermost call outward. Calls whose callee or arguments
failed before entry, calls with the wrong arity, and builtins do not add a user
function frame. The original error location is never replaced by an outer
statement's location. Syntax/runtime errors allow the REPL to continue;
input/output failures stop the session. Unfinished input at EOF is diagnosed by
the parser with its source location.

Locations use rune columns. ASCII source lines use a caret underline. Lines
containing non-ASCII text mark the offending range inline with `⟦` and `⟧`, for
example `"日本" ⟦+⟧ 1`. This works independently of terminal widths, including
emoji and combining sequences, and needs no external Unicode-width library.
An empty `⟦⟧` marks a zero-width position such as EOF. Tabs render as four spaces
and control characters are escaped. Multi-line ranges show their first line.
Hand-built ASTs without positions still produce message-only diagnostics.

The interpreter and its tests use only the Go standard library. Prefer the
standard library and justify any future external module by a concrete need.

## Feature
- [x] Unicode support
- [x] for-in loops
- [x] Syntax-template macros (`macro`, with ordinary call syntax)
- [x] logical operators

## Roadmap

### High Priority
- [x] Single-line comments (`//`)
- [ ] Block comments (`/* */`)
- [ ] Float type
- [x] Logical operators (`&&`, `||`)
- [x] Comparison operators (`<=`, `>=`)
- [x] Modulo operator (`%`)
- [x] `for-in` loops
- [ ] `while` loops + `break` / `continue`
- [x] Variable reassignment (`x = 10`)
- [ ] `type()` function

### Medium Priority
- [ ] Escape sequences (`\n`, `\t`, `\\`)
- [ ] Compound assignment operators (`+=`, `-=`, `*=`, `/=`)
- [ ] String functions (split, replace, trim, upper/lower)
- [ ] Array/Hash functions (map, filter, sort, reverse, keys, values)
- [ ] Type conversion functions (`int()`, `string()`, `float()`)
- [ ] String indexing (`str[0]`)

### Language design (future)
- [ ] Methods on structs without class inheritance (Go/Rust)
- [ ] Structural interfaces and explicit traits (`impl`-style)
- [ ] Function decorators as composition, not AST rewrites

### Low Priority
- [ ] switch / case
- [ ] File I/O
- [ ] Error handling (try/catch)
- [ ] import / module system
- [ ] Standard input (`input()`)

# Reference

## main text

- [Writing An Interpreter In Go](https://interpreterbook.com/)
- [The Lost Chapter: A Macro System For Monkey](https://interpreterbook.com/lost/)
- [Writing A Compiler In Go](https://compilerbook.com/)
- [monkey.org](https://monkeylang.org/)
- [official-code](https://interpreterbook.com/waiig_code_1.7.zip)

## subtext
- [kitasuke/monkey-go](https://github.com/kitasuke/monkey-go)
- [geovanisouza92/geo](https://github.com/geovanisouza92/geo)
- [skx/monkey](https://github.com/skx/monkey)
- [goby-lang/goby](https://github.com/goby-lang/goby)
- [prologic/monkey-lang](https://git.mills.io/prologic/monkey-lang)
- [newenclave/mico](https://github.com/newenclave/mico)
- [oldjun/pi](https://github.com/oldjun/pi)
- [ghost-language/ghost](https://github.com/ghost-language/ghost)
- [chrispine/crisp](https://github.com/chrispine/crisp)
- [langur](https://langurlang.org/)
- `*`[ludwigLanguage/ludwig](https://github.com/ludwigLanguage/ludwig)
- `*`[Flipez/rocket-lang](https://github.com/Flipez/rocket-lang)

`*` mean to be careful with license what uncertain or rigid.
