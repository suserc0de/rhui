# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-02

### Added
- Phase 0: Project bootstrap, repository layout, position tracker (`ast.Position`), and visual error diagnostics (`rhui.TemplateError`).
- Phase 1: High-performance state-machine Lexer/Tokenizer (`parser/lexer.go`, `parser/token.go`) with full `.rhui` syntax support.
- Phase 2: AST node hierarchy (`ast/node.go`) and Recursive Descent Parser (`parser/parser.go`) supporting HTML elements, components, conditionals (`<if>/<else>`), loops (`<each>`), and slots.
- Phase 3: Hierarchical execution scope (`runtime/scope.go`) and sandboxed expression evaluator (`runtime/evaluator.go`) supporting literals, field access, logical & comparison operators, and safe builtins.
- Phase 4: Component metadata registry (`compiler/component.go`), streaming renderer (`runtime/renderer.go`) with boolean attribute awareness, slot resolution, and top-level public engine API (`engine.go`).
- Phase 5: Contextual HTML escaping, XSS security test suite, and concurrency benchmark suite.
- Phase 6: Multi-framework adapters (Gin, Echo, Fiber, Chi, net/http) and Realtime Server-Sent Events package (`sse/`).
- Phase 7: Real-world admin dashboard application demo (`examples/admin-dashboard/`) and official VS Code syntax extension (`tooling/vscode/`).
- Enhancements: `embed.FS` support via `loader.Loader`, `devMode` auto-reload, buffer pooling via `sync.Pool`, `RenderWithLayout` helper, and `has_slot("name")` evaluator built-in function.
- Phase 8: Slot fallback content resolution and loop iteration metadata variables (`@index`, `@count`, `@first`, `@last`, `@even`, `@odd`).
- Phase 9: Pipe filter expressions (`|`) with built-in formatters (`upper`, `lower`, `trim`, `dateFormat`, `rupiah`, `default`, `json`) and `engine.RegisterFilter`.
- Phase 10: First-class HTMX fragment rendering (`engine.RenderFragment(w, name, fragmentID, data)`).
- Phase 11: Dynamic attribute spreading (`:attrs="extra_attrs"`) for generic wrapper components.
- Phase 12: First-class SEO package (`seo/`) with XML Sitemap generator and Schema.org JSON-LD structured data generator.
- Quality & CI: Multi-OS GitHub Actions CI workflow (`.github/workflows/ci.yml`) and Fuzz testing suite (`tests/parser_fuzz_test.go`).
- Documentation: Comprehensive `README.md` with official release badges, PayPal donation support, `LICENSE` (MIT), and `CONTRIBUTING.md`.
