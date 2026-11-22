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
a < b  // false
a > b  // true
a == b // false
a != b // true
```

Integer arithmetic uses signed 64-bit values. Division by zero and overflow
(in addition, subtraction, multiplication, division, or unary minus) return
a Monkey runtime error rather than panicking or silently wrapping. Division
truncates toward zero.

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

## Feature
- [x] Unicode support
- [x] for-in loops
- [x] logical operators

## Roadmap

### High Priority
- [x] Single-line comments (`//`)
- [ ] Block comments (`/* */`)
- [ ] Float type
- [x] Logical operators (`&&`, `||`)
- [ ] Comparison operators (`<=`, `>=`)
- [ ] Modulo operator (`%`)
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

### Low Priority
- [ ] switch / case
- [ ] File I/O
- [ ] Error handling (try/catch)
- [ ] import / module system
- [ ] Standard input (`input()`)

# Reference

## main text

- [Writing An Interpreter In Go](https://interpreterbook.com/)
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
