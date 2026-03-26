package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

// HTMLToMarkdown converts an HTML string to Markdown.
func HTMLToMarkdown(html string) (string, error) {
	md, err := htmltomarkdown.ConvertString(html)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(md), nil
}

// IsHTMLContentType returns true if the content type header indicates HTML.
func IsHTMLContentType(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "html")
}

func WriteMarkdownFile(fileName string, md string) {
	// Step 1: Ensure the directory exists
	err := os.MkdirAll("/home/dpauley/Documents/Code/Apps/Go/scraper_server/md", 0755)
	if err != nil {
		log.Printf("DIRECTORY ERROR: %s", err.Error())
	}

	// Step 2: Create the file
	file, err := os.Create(fmt.Sprintf("/home/dpauley/Documents/Code/Apps/Go/scraper_server/md/%s.md", fileName))
	if err != nil {
		log.Printf("FILE CREATE ERROR: %s", err.Error())
	}
	defer file.Close()

	// Step 3: Write to the file
	_, err = file.WriteString(md)
	if err != nil {
		log.Printf("MARKDOWN WRITE DOCUMENT ERROR: %s", err.Error())
	}
}
