package tests

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/coderhaqi/rhui"
	"github.com/coderhaqi/rhui/seo"
)

func TestSlotFallbackContent(t *testing.T) {
	engine, err := rhui.New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	cardTmpl := `<component name="Notice">
  <div class="notice">
    <slot>
      <span class="default-msg">No notices at this time.</span>
    </slot>
  </div>
</component>`

	_ = engine.ParseTemplate("components/Notice", cardTmpl)

	// Case 1: Caller leaves slot empty -> Fallback rendered
	page1 := `<ui:Notice />`
	_ = engine.ParseTemplate("page1", page1)
	out1, _ := engine.RenderString("page1", nil)
	if !strings.Contains(out1, "No notices at this time.") {
		t.Fatalf("expected fallback slot content rendered, got: %s", out1)
	}

	// Case 2: Caller provides content -> Fallback replaced
	page2 := `<ui:Notice><p>Urgent Notice!</p></ui:Notice>`
	_ = engine.ParseTemplate("page2", page2)
	out2, _ := engine.RenderString("page2", nil)
	if strings.Contains(out2, "No notices at this time.") {
		t.Fatalf("expected fallback slot to be overridden, got: %s", out2)
	}
	if !strings.Contains(out2, "<p>Urgent Notice!</p>") {
		t.Fatalf("expected caller slot content, got: %s", out2)
	}
}

func TestLoopContextMetadata(t *testing.T) {
	engine, _ := rhui.New()

	tmpl := `<each items="Items" as="it">
  <span>Index:{{ @index }}, Count:{{ @count }}, First:{{ @first }}, Last:{{ @last }}, Even:{{ @even }}</span>
</each>`

	_ = engine.ParseTemplate("loop-test", tmpl)

	out, err := engine.RenderString("loop-test", map[string]any{
		"Items": []string{"Alpha", "Beta", "Gamma"},
	})
	if err != nil {
		t.Fatalf("render loop failed: %v", err)
	}

	if !strings.Contains(out, "Index:0, Count:1, First:true, Last:false, Even:false") {
		t.Errorf("missing correct first iteration meta: %s", out)
	}
	if !strings.Contains(out, "Index:1, Count:2, First:false, Last:false, Even:true") {
		t.Errorf("missing correct second iteration meta: %s", out)
	}
	if !strings.Contains(out, "Index:2, Count:3, First:false, Last:true, Even:false") {
		t.Errorf("missing correct last iteration meta: %s", out)
	}
}

func TestPipeFilters(t *testing.T) {
	engine, _ := rhui.New()

	// Register custom filter
	engine.RegisterFilter("shout", func(val any) string {
		return strings.ToUpper(val.(string)) + "!!!"
	})

	tmpl := `<div>
  <p>{{ Name | upper }}</p>
  <p>{{ Name | shout }}</p>
  <p>{{ Price | rupiah }}</p>
  <p>{{ Empty | default("Default Value") }}</p>
</div>`

	_ = engine.ParseTemplate("pipe-test", tmpl)

	out, err := engine.RenderString("pipe-test", map[string]any{
		"Name":  "raden haqi",
		"Price": 1500000,
		"Empty": "",
	})
	if err != nil {
		t.Fatalf("render pipe test failed: %v", err)
	}

	if !strings.Contains(out, "<p>RADEN HAQI</p>") {
		t.Errorf("upper filter failed: %s", out)
	}
	if !strings.Contains(out, "<p>RADEN HAQI!!!</p>") {
		t.Errorf("custom shout filter failed: %s", out)
	}
	if !strings.Contains(out, "<p>Rp 1.500.000</p>") {
		t.Errorf("rupiah filter failed: %s", out)
	}
	if !strings.Contains(out, "<p>Default Value</p>") {
		t.Errorf("default fallback filter failed: %s", out)
	}
}

func TestRenderFragment(t *testing.T) {
	engine, _ := rhui.New()

	pageTmpl := `<div class="container">
  <h1>Page Title</h1>
  <table fragment="user-rows">
    <tr><td>Row 1</td></tr>
    <tr><td>Row 2</td></tr>
  </table>
  <footer>Footer info</footer>
</div>`

	_ = engine.ParseTemplate("users", pageTmpl)

	var buf bytes.Buffer
	err := engine.RenderFragment(&buf, "users", "user-rows", nil)
	if err != nil {
		t.Fatalf("RenderFragment failed: %v", err)
	}

	res := buf.String()
	if !strings.HasPrefix(res, "<table") || !strings.Contains(res, "Row 1") {
		t.Errorf("expected only table fragment, got: %s", res)
	}
	if strings.Contains(res, "Page Title") || strings.Contains(res, "Footer info") {
		t.Errorf("RenderFragment leaked outside content: %s", res)
	}
}

func TestDynamicAttributeSpreading(t *testing.T) {
	engine, _ := rhui.New()

	btnTmpl := `<component name="IconButton">
  <props>
    <prop name="icon" default="star" />
  </props>
  <button class="btn" :attrs="extra_attrs">
    <i class="icon-{{ icon }}"></i>
    <slot />
  </button>
</component>`

	_ = engine.ParseTemplate("components/IconButton", btnTmpl)

	pageTmpl := `<ui:IconButton icon="heart" id="favorite-btn" hx-post="/fav" data-role="action">Like</ui:IconButton>`
	_ = engine.ParseTemplate("page", pageTmpl)

	out, err := engine.RenderString("page", nil)
	if err != nil {
		t.Fatalf("render attribute spread failed: %v", err)
	}

	if !strings.Contains(out, "id=\"favorite-btn\"") {
		t.Errorf("missing spread id attribute: %s", out)
	}
	if !strings.Contains(out, "hx-post=\"/fav\"") {
		t.Errorf("missing spread hx-post attribute: %s", out)
	}
	if !strings.Contains(out, "data-role=\"action\"") {
		t.Errorf("missing spread data-role attribute: %s", out)
	}
}

func TestSEOSitemapAndJSONLD(t *testing.T) {
	sitemap := seo.NewSitemap("https://coderhaqi.com")
	now := time.Now()
	sitemap.Add(seo.URL{
		Loc:        "/articles/go-template-engine",
		LastMod:    &now,
		ChangeFreq: "daily",
		Priority:   0.9,
	})

	var buf bytes.Buffer
	err := sitemap.Render(&buf)
	if err != nil {
		t.Fatalf("sitemap render failed: %v", err)
	}

	sitemapXML := buf.String()
	if !strings.Contains(sitemapXML, "<loc>https://coderhaqi.com/articles/go-template-engine</loc>") {
		t.Errorf("sitemap missing full URL: %s", sitemapXML)
	}

	// JSON-LD test
	schema := map[string]any{
		"@context": "https://schema.org",
		"@type":    "Article",
		"headline": "Go RHUI Template Engine",
	}
	jsonldTag, err := seo.JSONLD(schema)
	if err != nil {
		t.Fatalf("jsonld generation failed: %v", err)
	}

	if !strings.Contains(jsonldTag, "<script type=\"application/ld+json\">") || !strings.Contains(jsonldTag, "Go RHUI Template Engine") {
		t.Errorf("invalid jsonld output: %s", jsonldTag)
	}
}
