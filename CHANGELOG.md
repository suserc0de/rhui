# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.8] - 2026-11-08

### API Ergonomics & Documentation (P0)
- **Idiomatic Constructor (`rhui.New`)**: Standardized `rhui.New(opts ...Option) (*Engine, error)` as the sole primary constructor for idiomatic error handling.
- **Convenience Helper (`rhui.Must`)**: Added `rhui.Must(e *Engine, err error) *Engine` for panic-on-failure package-level initializations matching standard library conventions (`template.Must`).
- **Deprecated Panic-Prone Constructor (`rhui.NewEngine`)**: Formally marked `rhui.NewEngine()` and `rhui.EngineOptions` as deprecated, re-routing `NewEngine()` safely to `Must(New(opts...))` with clear GoDoc migration guidance.
- **Comprehensive GoDoc & Runnable Examples**: Added comprehensive documentation and runnable GoDoc examples (`Example_basic`, `ExampleMust`, `ExampleWithFS`, `ExampleEngine_RenderFragment`, `ExampleEngine_RegisterFilter`, `ExampleEngine_RegisterFunc`, `ExampleEngine_RenderWithLayout`) verified via `go test`.

### Benchmark Credibility & 100% Parity (P0)
- **100% Identical Head-to-Head Benchmark**: Standardized benchmark test cases between RHUI and `html/template` with 100% identical data inputs, DOM structures, and output formatting.
- **Automated Output Parity Verification**: Added `TestBenchmarkOutputParity` to guarantee 100% semantic HTML equality before executing performance benchmarks.

## [1.0.7] - 2026-11-07

### CSS Framework Ergonomics & Attributes (P0)
- **Attribute Value Interpolation**: Evaluates expressions embedded inside static attributes (`class="alert alert-{{ type }} {{ customClass }}"`, `id="box-{{ id }}"`, `data-role="{{ role }}"`). Precompiled into AST segments at parse-time for maximum rendering speed and escaped via `EscapeAttribute()`.
- **Dynamic `:class` Binding**: Added support for map and slice class expressions (`:class="{'btn-primary': isPrimary, 'opacity-50': disabled}"` or `:class="['btn', customClass]"`).
- **Dynamic `:style` Binding**: Added support for inline CSS style maps (`:style="{'background-color': color, 'font-size': size + 'px'}"`).
- **Inline Map & Slice Literals in Expressions**: Parser now supports `{ 'key': val }` and `[ 'item1', 'item2' ]` syntax directly within template expressions.
- **Tailwind CSS & Bootstrap Component Kits**: Added comprehensive verification and full compatibility test suites for Tailwind CSS utility composition and Bootstrap 5 components.

## [1.0.5] - 2026-11-06

### Security Hardening (P0)
- **Strict URL Scheme Allowlist**: Transitioned `SanitizeURL` from blocklisting to strict allowlisting (`http://`, `https://`, `mailto:`, `tel:`, `sms:`, and relative/fragment paths `#`, `/`, `./`, `../`, `?`). Neutralized all pseudo-protocols (`javascript:`, `vbscript:`, `data:`, `blob:`, `file:`) and protocol-relative schemes (`//`).
- **SSE CRLF Header & Event Injection Neutralization**: Sanitized carriage return (`\r`) and newline (`\n`) characters in SSE `EventName` and `ID` to prevent event frame tampering.
- **Configurable SSE CORS Origin**: Added `sse.WithAllowedOrigin` option to restrict Server-Sent Events stream origins instead of mandatory wildcard (`*`).

### Engine Resilience & DoS Prevention (P0)
- **Bounded LRU Expression Cache**: Replaced unbounded map with a high-performance, thread-safe LRU Cache (`expr.DefaultLRUCapacity = 4096`) for expression parsing, preventing memory exhaustion and OOM attacks.
- **Configurable Component Recursion Guard**: Implemented recursion depth checking (`WithMaxRenderDepth`, default: 64) to prevent stack overflows and CPU starvation from circular component references.

## [1.0.4] - 2026-11-03

### Security (P0)
- **Inline Event Handler Blocking**: Blocked all dangerous inline event handlers (`onclick`, `onerror`, `onload`, `onmouseover`, etc.) from being injected via `:attrs` dynamic attribute spreading.
- **Negative Fuzzing Coverage**: Verified 50+ DOM Level 2/3 and HTML5 event handler attributes are strictly discarded.

### Concurrency (P0)
- **Atomic Snapshotting (`Reload` vs `Render`)**: Replaced mutable template and renderer pointers with atomic engine snapshot storage (`atomic.Pointer[engineState]`), completely eliminating data races during background template reloads under heavy concurrent read traffic.

### API & DX (P1)
- **Fixed `NewEngine.AutoReload`**: Ensured `EngineOptions.AutoReload` properly configures dev mode (`WithDevMode(true)`).

### Integration Testing (P1)
- Added full-stack integration tests for Gin, Echo, Fiber, Chi, and standard `net/http` routers.
- Added concurrent multi-client SSE broadcast verification (20 simultaneous active listeners).
- Added multi-threaded concurrent reload stress testing (0 data race verified via `go test -race`).

## [1.0.3] - 2026-11-03

### Security Hardening (P0)
- **Dynamic Attribute Validation**: Implemented strict W3C HTML5 attribute name validation (`IsValidAttributeName`) for dynamic attribute spreading (`:attrs`), rejecting space/quote breakout attacks.
- **Obfuscated URL Scheme Neutralization**: Hardened `SanitizeURL` against null-byte, control character, and tab-separated obfuscated URI schemes (`java\0script:`, `data:image/svg+xml`).
- **Penetration Fuzzing Suite**: Added continuous automated fuzz tests (`FuzzXSSPenetration`) verifying zero breakout vulnerabilities.

### Memory & Scope Safety (P0)
- **Error-Safe Lifecycle Recycling**: Enforced `defer ReleaseScope` across root rendering and component invocations to guarantee zero leaks even on render failure.
- **Continuous Memory Leak Test**: Verified 100,000 iterations of nested component rendering maintain bounded heap allocation without linear growth.

### Compile-Time Optimizations (P1)
- **Parse-Time Expression AST Embedding**: Embedded pre-compiled `expr.Expr` directly into `ast.InterpolationNode`, `ast.IfNode`, `ast.EachNode`, and `ast.Attribute`, eliminating global lock contention and map lookups at runtime.
- **2x-3x Speedup**: Reduced single component rendering to **1.29 µs** (from 4.15 µs in v1.0.1) and complex 50-row rendering to **156 µs** (beating stdlib `html/template`).

### Benchmarks (P1)
- Added 100-row, 1,000-row table scaling benchmarks, nested component benchmark, and parallel goroutine concurrency benchmarks.

## [1.0.2] - 2026-11-03

### Performance Optimizations
- **Compiled Expression AST**: Introduced thread-safe expression AST parsing and compilation (`expr.Expr`) with caching and pooled evaluation contexts (`sync.Pool`), eliminating redundant token parsing during runtime render loops.
- **Fast-Path Zero-Allocation Escaping**: Implemented sensitive character byte scanning (`strings.IndexAny`) in `EscapeHTML` and `EscapeAttribute` to bypass heap allocations for plain text.
- **Direct Writer Streaming**: Replaced all `fmt.Fprintf` calls in `renderElement` with zero-allocation direct string/slice writes (`io.WriteString`) and fast primitive formatting (`strconv`).
- **Loop Sub-Scope Recycling**: Recycled `Scope` instances in `<each>` loop iterations using `sync.Pool`, significantly reducing GC pressure on large tables.

## [1.0.1] - 2026-11-03

### Fixed
- Fixed documentation links, module path consistency, and reflected dynamic map/struct lookups in runtime scope.

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
