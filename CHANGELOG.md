# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Phase 0: Project bootstrap, repository layout, position tracker (`ast.Position`), and visual error diagnostics (`rhui.TemplateError`).
- Phase 1: High-performance state-machine Lexer/Tokenizer (`parser/lexer.go`, `parser/token.go`) with full `.rhui` syntax support.
- Phase 2: AST node hierarchy (`ast/node.go`) and Recursive Descent Parser (`parser/parser.go`) supporting HTML elements, components, conditionals (`<if>/<else>`), loops (`<each>`), and slots.
- Phase 3: Hierarchical execution scope (`runtime/scope.go`) and sandboxed expression evaluator (`runtime/evaluator.go`) supporting literals, field access, logical & comparison operators, and safe builtins.
- Phase 4: Component metadata registry (`compiler/component.go`), streaming renderer (`runtime/renderer.go`) with boolean attribute awareness, slot resolution, and top-level public engine API (`engine.go`).
- Phase 5: Contextual HTML escaping, XSS security test suite, and concurrency benchmark suite.
- Phase 6: Multi-framework adapters (Gin, Echo, Fiber, Chi/nethttp) and Realtime Server-Sent Events package (`sse/`).
- Documentation: Comprehensive `README.md`, `LICENSE` (MIT), and `CONTRIBUTING.md`.
