# RHUI Showcase: Realtime Store (HTMX + SSE + Tailwind)

A production-grade, zero-JavaScript-build demonstration showing how **RHUI** delivers modern reactive user experiences directly from Go standard library (`net/http`) without Node.js, Webpack, or client-side SPA frameworks.

---

## 🌟 Architectural Highlights

1. **Initial Full-Page Server-Side Rendering (SSR)**:
   - Clean HTML served with pre-compiled AST components (`<ui:ProductCard>`, `<ui:StockBadge>`).
   - Zero hydration lag and complete SEO friendliness.

2. **Targeted Partial Swaps via `RenderFragment` (HTMX)**:
   - When users click **Add to Cart**, HTMX issues a `POST /cart/add?id=...`.
   - The Go backend increments the cart counter and renders **only** the `<div fragment="cart-badge">` node using `engine.RenderFragment()`, swapping just the badge counter without re-rendering the layout.

3. **Live Remote Inventory Updates via Server-Sent Events (SSE)**:
   - Connects natively to `rhui/sse` broker (`sse-connect="/sse/live-stock"`).
   - When remote orders occur or inventory changes, the server broadcasts an updated `<ui:StockBadge>` component frame to all connected browsers in real-time.

---

## 🚀 Running the Showcase Locally

```bash
cd examples/realtime-store
go run .
```

Open your browser at:
```text
http://localhost:8081
```

1. Open two browser tabs side-by-side at `http://localhost:8081`.
2. Click **"Simulate Remote Customer Order"** in tab 1 &rarr; observe the stock badge in tab 2 update **instantly in real-time** over SSE.
3. Click **"Add to Cart"** &rarr; observe the cart counter update smoothly via targeted fragment swap.
