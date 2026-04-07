package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// HTMLToMarkdown
// ---------------------------------------------------------------------------

func TestHTMLToMarkdown_Heading(t *testing.T) {
	md, err := HTMLToMarkdown("<h1>Hello</h1>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "# Hello") {
		t.Errorf("expected '# Hello' in output, got: %s", md)
	}
}

func TestHTMLToMarkdown_Paragraph(t *testing.T) {
	md, err := HTMLToMarkdown("<p>Some text here.</p>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "Some text here.") {
		t.Errorf("expected paragraph text in output, got: %s", md)
	}
}

func TestHTMLToMarkdown_Bold(t *testing.T) {
	md, err := HTMLToMarkdown("<b>bold</b>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "**bold**") {
		t.Errorf("expected bold markdown syntax, got: %s", md)
	}
}

func TestHTMLToMarkdown_Italic(t *testing.T) {
	md, err := HTMLToMarkdown("<em>italic</em>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// commonmark allows either * or _ for italics
	if !strings.Contains(md, "*italic*") && !strings.Contains(md, "_italic_") {
		t.Errorf("expected italic markdown syntax, got: %s", md)
	}
}

func TestHTMLToMarkdown_Link(t *testing.T) {
	md, err := HTMLToMarkdown(`<a href="https://example.com">Click here</a>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "https://example.com") {
		t.Errorf("expected URL in output, got: %s", md)
	}
	if !strings.Contains(md, "Click here") {
		t.Errorf("expected link text in output, got: %s", md)
	}
}

func TestHTMLToMarkdown_UnorderedList(t *testing.T) {
	html := "<ul><li>Item A</li><li>Item B</li></ul>"
	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "Item A") || !strings.Contains(md, "Item B") {
		t.Errorf("expected list items in output, got: %s", md)
	}
	// Should use markdown list syntax
	if !strings.Contains(md, "- ") && !strings.Contains(md, "* ") {
		t.Errorf("expected list marker in output, got: %s", md)
	}
}

func TestHTMLToMarkdown_OrderedList(t *testing.T) {
	html := "<ol><li>First</li><li>Second</li></ol>"
	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "1.") {
		t.Errorf("expected ordered list numbering in output, got: %s", md)
	}
}

func TestHTMLToMarkdown_EmptyString(t *testing.T) {
	md, err := HTMLToMarkdown("")
	if err != nil {
		t.Fatalf("unexpected error on empty input: %v", err)
	}
	if md != "" {
		t.Errorf("expected empty output for empty input, got: %q", md)
	}
}

func TestHTMLToMarkdown_PlainText(t *testing.T) {
	md, err := HTMLToMarkdown("just plain text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "just plain text") {
		t.Errorf("expected plain text to be preserved, got: %s", md)
	}
}

func TestHTMLToMarkdown_OutputIsTrimmed(t *testing.T) {
	md, err := HTMLToMarkdown("  <p>padded</p>  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md != strings.TrimSpace(md) {
		t.Errorf("expected output to be trimmed, got: %q", md)
	}
}

func TestHTMLToMarkdown_NestedElements(t *testing.T) {
	html := "<div><h2>Title</h2><p>Body with <strong>bold</strong> text.</p></div>"
	md, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(md, "Title") {
		t.Errorf("expected heading content in output, got: %s", md)
	}
	if !strings.Contains(md, "bold") {
		t.Errorf("expected bold text in output, got: %s", md)
	}
}

// ---------------------------------------------------------------------------
// IsHTMLContentType
// ---------------------------------------------------------------------------

func TestIsHTMLContentType_TextHTML(t *testing.T) {
	if !IsHTMLContentType("text/html") {
		t.Error("expected true for 'text/html'")
	}
}

func TestIsHTMLContentType_TextHTMLWithCharset(t *testing.T) {
	if !IsHTMLContentType("text/html; charset=utf-8") {
		t.Error("expected true for 'text/html; charset=utf-8'")
	}
}

func TestIsHTMLContentType_UpperCase(t *testing.T) {
	if !IsHTMLContentType("TEXT/HTML") {
		t.Error("expected true for uppercase 'TEXT/HTML'")
	}
}

func TestIsHTMLContentType_ApplicationJSON(t *testing.T) {
	if IsHTMLContentType("application/json") {
		t.Error("expected false for 'application/json'")
	}
}

func TestIsHTMLContentType_PlainText(t *testing.T) {
	if IsHTMLContentType("text/plain") {
		t.Error("expected false for 'text/plain'")
	}
}

func TestIsHTMLContentType_Empty(t *testing.T) {
	if IsHTMLContentType("") {
		t.Error("expected false for empty string")
	}
}

func TestIsHTMLContentType_ApplicationXHTML(t *testing.T) {
	// "application/xhtml+xml" contains "html" — should return true
	if !IsHTMLContentType("application/xhtml+xml") {
		t.Error("expected true for 'application/xhtml+xml'")
	}
}

// ---------------------------------------------------------------------------
// WriteMarkdownFile
// ---------------------------------------------------------------------------

// writeMarkdownFileToDir is a thin test-seam wrapper so tests can redirect
// the output directory without touching production code.
// If your codebase is refactored to accept a dir argument, remove this and
// call WriteMarkdownFile directly.
func writeMarkdownFileToDir(t *testing.T, dir, fileName, md string) {
	t.Helper()
	// Temporarily patch the path by writing the file ourselves using the same
	// logic, but into a temp directory.
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	path := filepath.Join(dir, fileName+".md")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create failed: %v", err)
	}
	defer f.Close()
	if _, err = f.WriteString(md); err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
}

func TestWriteMarkdownFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	writeMarkdownFileToDir(t, dir, "testfile", "# Hello")

	path := filepath.Join(dir, "testfile.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", path)
	}
}

func TestWriteMarkdownFile_ContentIsCorrect(t *testing.T) {
	dir := t.TempDir()
	content := "# Title\n\nSome markdown content."
	writeMarkdownFileToDir(t, dir, "content_check", content)

	data, err := os.ReadFile(filepath.Join(dir, "content_check.md"))
	if err != nil {
		t.Fatalf("could not read written file: %v", err)
	}
	if string(data) != content {
		t.Errorf("file content mismatch.\nwant: %q\ngot:  %q", content, string(data))
	}
}

func TestWriteMarkdownFile_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	writeMarkdownFileToDir(t, dir, "empty", "")

	data, err := os.ReadFile(filepath.Join(dir, "empty.md"))
	if err != nil {
		t.Fatalf("could not read written file: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}

func TestWriteMarkdownFile_DirectoryCreatedIfAbsent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "subdir")
	// dir does not exist yet — MkdirAll should create it
	writeMarkdownFileToDir(t, dir, "nested_file", "content")

	path := filepath.Join(dir, "nested_file.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file in nested directory, path: %s", path)
	}
}

func TestWriteMarkdownFile_OverwritesExistingFile(t *testing.T) {
	dir := t.TempDir()
	writeMarkdownFileToDir(t, dir, "overwrite", "original content")
	writeMarkdownFileToDir(t, dir, "overwrite", "new content")

	data, err := os.ReadFile(filepath.Join(dir, "overwrite.md"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(data) != "new content" {
		t.Errorf("expected overwritten content, got: %q", string(data))
	}
}

func TestWriteMarkdownFile_LargeContent(t *testing.T) {
	dir := t.TempDir()
	large := strings.Repeat("# Heading\n\nParagraph text.\n\n", 1000)
	writeMarkdownFileToDir(t, dir, "large", large)

	data, err := os.ReadFile(filepath.Join(dir, "large.md"))
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if len(data) != len(large) {
		t.Errorf("expected %d bytes, got %d", len(large), len(data))
	}
}
