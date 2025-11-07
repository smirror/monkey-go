## Summary

This PR adds a comprehensive feature roadmap document (`ROADMAP.md`) that outlines planned language extensions for monkey-go, with a focus on three key design goals:

1. **Turing Completeness** - Making the language practically useful with loops and mutable state
2. **Type Safety** - Introducing optional type annotations, null safety, and error handling
3. **Developer Experience** - Adding modern language features for better ergonomics

## Changes

- ✨ **New**: `ROADMAP.md` - Comprehensive feature roadmap with 7 phases of development
- 📝 **Modified**: `README.md` - Added link to the roadmap document

## Roadmap Highlights

### Phase 1: Practical Turing Completeness 🔴 Highest Priority
- Variable reassignment
- Logical operators (`&&`, `||`) ⭐ README TODO
- `while` loops
- `for` loops ⭐ README TODO
- `break`/`continue` statements

### Phase 2: Basic Type Safety 🟠
- Enhanced runtime type checking
- Optional type annotations
- Null safety (Option types)
- Result types for error handling

### Phase 3: Improved Ergonomics 🟡
- Complete comparison operators (`>=`, `<=`)
- Arithmetic operators (`%`, `**`, `+=`, `++`)
- Floating-point numbers
- Pattern matching
- Destructuring
- Template strings

### Phase 4-7: Advanced Features 🟢🔵
- Standard library expansion
- Module system
- Type inference & generics
- Compiler implementation (bytecode VM)

## Context

The original "Writing An Interpreter In Go" book intentionally excludes loops and logical operators, relying on recursion and higher-order functions for iteration. This roadmap extends the language with practical features while maintaining backward compatibility.

## Implementation Priority

The roadmap prioritizes "Quick Wins" that provide immediate value:
1. Logical operators (`&&`, `||`) - 1-2 days
2. Comparison operators (`>=`, `<=`) - 1 day
3. Modulo operator (`%`) - 1 day
4. Variable reassignment - 2-3 days

These features alone would significantly improve the language's usability.

## Documentation

Each feature in the roadmap includes:
- Current status and target goal
- Code examples
- Implementation locations in the codebase
- Priority level (🔴🟠🟡🟢🔵)
- Difficulty estimate (⭐1-5)
- Dependencies

## References

- [Writing An Interpreter In Go](https://interpreterbook.com/)
- [Writing A Compiler In Go](https://compilerbook.com/)
- Extended implementations: [skx/monkey](https://github.com/skx/monkey), [bradford-hamilton/monkey-lang](https://github.com/bradford-hamilton/monkey-lang)

## Test Plan

- [x] Document is well-structured and readable
- [x] README link to roadmap works correctly
- [x] All priority levels and dependencies are clearly marked
- [x] Code examples are valid Monkey syntax
