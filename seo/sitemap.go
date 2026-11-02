package seo

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// URL represents a sitemap item.
type URL struct {
	Loc        string     `xml:"loc"`
	LastMod    *time.Time `xml:"lastmod,omitempty"`
	ChangeFreq string     `xml:"changefreq,omitempty"`
	Priority   float32    `xml:"priority,omitempty"`
}

// URLSet is the root container of an XML sitemap.
type URLSet struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL    `xml:"url"`
}

// Sitemap generates valid XML sitemaps.
type Sitemap struct {
	baseURL string
	urls    []URL
}

// NewSitemap creates a new Sitemap instance with a base URL.
func NewSitemap(baseURL string) *Sitemap {
	return &Sitemap{
		baseURL: baseURL,
	}
}

// Add adds a URL to the sitemap.
func (s *Sitemap) Add(u URL) {
	if s.baseURL != "" && len(u.Loc) > 0 && u.Loc[0] == '/' {
		u.Loc = s.baseURL + u.Loc
	}
	s.urls = append(s.urls, u)
}

// Render writes the XML sitemap to an io.Writer.
func (s *Sitemap) Render(w io.Writer) error {
	urlset := URLSet{URLs: s.urls}
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	return enc.Encode(urlset)
}

// ServeHTTP serves the sitemap over HTTP with correct headers.
func (s *Sitemap) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_ = s.Render(w)
}

// JSONLD returns an HTML safe <script type="application/ld+json"> tag for structured Schema.org data.
func JSONLD(schema any) (string, error) {
	bytes, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("<script type=\"application/ld+json\">%s</script>", string(bytes)), nil
}
