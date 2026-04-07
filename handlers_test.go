package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newRequest builds a GET request whose URL query param points at targetURL.
// Adjust the param name to match your actual GetURL(r) implementation.
func newRequest(t *testing.T, targetURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "/?url="+targetURL, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	return req
}

// startFakeServer spins up an httptest.Server that serves the provided HTML body.
func startFakeServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
}

// ---------------------------------------------------------------------------
// MarkdownHandler
// ---------------------------------------------------------------------------

func TestMarkdownHandler_ValidHTML(t *testing.T) {
	html := `<html><head></head><body><h1>Hello World</h1><p>Some text.</p></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	MarkdownHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result MarkdownElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.Content == "" {
		t.Error("expected non-empty markdown content")
	}
	if !strings.Contains(result.Content, "Hello World") {
		t.Errorf("markdown should contain heading text, got: %s", result.Content)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode 200 in payload, got %d", result.StatusCode)
	}
	if result.Method != http.MethodGet {
		t.Errorf("expected Method GET in payload, got %s", result.Method)
	}
}

func TestMarkdownHandler_NonHTMLResponse(t *testing.T) {
	// Server returns plain JSON — not HTML
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"value"}`))
	}))
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	MarkdownHandler(rr, req)

	// Handler should return early; body should be empty or contain an error,
	// NOT a valid MarkdownElement with content.
	var result MarkdownElement
	_ = json.NewDecoder(rr.Body).Decode(&result)
	if result.Content != "" {
		t.Errorf("expected no markdown content for non-HTML response, got: %s", result.Content)
	}
}

func TestMarkdownHandler_UnreachableURL(t *testing.T) {
	// Point at a port that is not listening
	req := newRequest(t, "http://127.0.0.1:19999")
	rr := httptest.NewRecorder()

	// Should not panic; handler must deal with the GET error gracefully
	MarkdownHandler(rr, req)

	var result MarkdownElement
	_ = json.NewDecoder(rr.Body).Decode(&result)
	if result.Content != "" {
		t.Error("expected no content for unreachable URL")
	}
}

// ---------------------------------------------------------------------------
// LinkHandler
// ---------------------------------------------------------------------------

func TestLinkHandler_ExtractsLinks(t *testing.T) {
	html := `<html><body>
		<a href="/page1">Page 1</a>
		<a href="/page2">Page 2</a>
		<a href="https://external.example.com">External</a>
	</body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	LinkHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result LinkElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if len(result.Links) == 0 {
		t.Fatal("expected links to be extracted")
	}

	// Relative links should be resolved to absolute
	for link := range result.Links {
		if !strings.HasPrefix(link, "http") {
			t.Errorf("all links should be absolute, got: %s", link)
		}
	}

	// Verify specific absolute path is present
	found := false
	for link := range result.Links {
		if strings.HasSuffix(link, "/page1") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected /page1 link to be resolved and present, links: %v", result.Links)
	}
}

func TestLinkHandler_NoLinks(t *testing.T) {
	html := `<html><body><p>No links here.</p></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	LinkHandler(rr, req)

	var result LinkElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if len(result.Links) != 0 {
		t.Errorf("expected 0 links, got %d", len(result.Links))
	}
}

func TestLinkHandler_DuplicateLinksAreCounted(t *testing.T) {
	html := `<html><body>
		<a href="/dup">Link A</a>
		<a href="/dup">Link B</a>
	</body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	LinkHandler(rr, req)

	var result LinkElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	for link, count := range result.Links {
		if strings.HasSuffix(link, "/dup") && count != 2 {
			t.Errorf("expected duplicate link count of 2, got %d", count)
		}
	}
}

func TestLinkHandler_StatusCodePopulated(t *testing.T) {
	html := `<html><body><a href="/x">X</a></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	LinkHandler(rr, req)

	var result LinkElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode 200, got %d", result.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// TextHandler
// ---------------------------------------------------------------------------

func TestTextHandler_ExtractsBodyText(t *testing.T) {
	html := `<html><body><p>Hello, scraper!</p><span>More text.</span></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	TextHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result PageElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.Text == "" {
		t.Error("expected non-empty body text")
	}
	if !strings.Contains(result.Text, "Hello, scraper!") {
		t.Errorf("expected extracted text to contain page content, got: %s", result.Text)
	}
}

func TestTextHandler_EmptyBody(t *testing.T) {
	html := `<html><body></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	TextHandler(rr, req)

	var result PageElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if strings.TrimSpace(result.Text) != "" {
		t.Errorf("expected empty text for empty body, got: %q", result.Text)
	}
}

func TestTextHandler_MetadataPopulated(t *testing.T) {
	html := `<html><body><p>text</p></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	TextHandler(rr, req)

	var result PageElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode 200, got %d", result.StatusCode)
	}
	if result.Method != http.MethodGet {
		t.Errorf("expected Method GET, got %s", result.Method)
	}
}

// ---------------------------------------------------------------------------
// TableHandler
// ---------------------------------------------------------------------------

func TestTableHandler_ExtractsTableText(t *testing.T) {
	html := `<html><body><table>
		<tr><td>Name</td><td>Age</td></tr>
		<tr><td>Alice</td><td>30</td></tr>
		<tr><td>Bob</td><td>25</td></tr>
	</table></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	TableHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result TableElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.TableText == "" {
		t.Error("expected non-empty table text")
	}
	if !strings.Contains(result.TableText, "Alice") {
		t.Errorf("expected table text to contain cell data, got: %s", result.TableText)
	}
	if !strings.Contains(result.TableText, "Name") {
		t.Errorf("expected table text to contain header data, got: %s", result.TableText)
	}
}

func TestTableHandler_NoTable(t *testing.T) {
	html := `<html><body><p>No table here.</p></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	TableHandler(rr, req)

	var result TableElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.TableText != "" {
		t.Errorf("expected empty TableText for page with no table, got: %s", result.TableText)
	}
}

func TestTableHandler_MultipleRows(t *testing.T) {
	rows := 50
	var sb strings.Builder
	sb.WriteString("<html><body><table>")
	for i := 0; i < rows; i++ {
		sb.WriteString("<tr><td>Row</td></tr>")
	}
	sb.WriteString("</table></body></html>")

	srv := startFakeServer(t, sb.String())
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	TableHandler(rr, req)

	var result TableElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	occurrences := strings.Count(result.TableText, "Row")
	if occurrences != rows {
		t.Errorf("expected %d row occurrences, got %d", rows, occurrences)
	}
}

func TestTableHandler_StatusCodePopulated(t *testing.T) {
	html := `<html><body><table><tr><td>x</td></tr></table></body></html>`
	srv := startFakeServer(t, html)
	defer srv.Close()

	req := newRequest(t, srv.URL)
	rr := httptest.NewRecorder()

	TableHandler(rr, req)

	var result TableElement
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Errorf("expected StatusCode 200, got %d", result.StatusCode)
	}
}
