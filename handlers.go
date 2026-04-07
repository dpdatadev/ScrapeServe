package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Convert all text on page to Markdown doc
func MarkdownHandler(w http.ResponseWriter, r *http.Request) {
	var fileName string // Markdown will be written to disk
	URL := GetURL(r)
	// Get request to URL
	response, err := http.Get(URL)
	if err != nil {
		log.Println("http GET error:", errors.New(err.Error()))
		return
	}

	defer response.Body.Close()

	// Load bytes of request to website
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("io error:", errors.New(err.Error()))
		return
	}

	// Markdown object to hold converted html
	m := &MarkdownElement{}
	// Convert the bytes to string
	tryHtmlString := string(body)
	// Ensure the new string actually contains HTML, err if it doesn't
	if IsHTMLContentType(tryHtmlString) {
		m.Content, err = HTMLToMarkdown(tryHtmlString)
		if err != nil {
			log.Println("error:", errors.New(err.Error()))
		}
	} else {
		log.Println("REQUEST MUST BE HTML TO CONVERT TO MARKDOWN")
		log.Println(tryHtmlString)
		return
	}
	// Assign default metadata
	m.StatusCode = 200
	m.Method = r.Method
	m.Host = r.Host
	// Track the request
	tagString := strconv.Itoa(m.StatusCode) + "::" + m.Method
	log.Printf("MARKDOWN ReqMETA (url: %s): %s", URL, tagString)
	//Write md document
	fileName = strings.ReplaceAll(URL, "/", "")
	fileName = strings.ReplaceAll(fileName, ".com", "")
	fileName = strings.ReplaceAll(fileName, ".", "")
	fileName = strings.ReplaceAll(fileName, ":", "")
	fileName = strings.ReplaceAll(fileName, "www", "")
	fileName = strings.ReplaceAll(fileName, "https", "")
	WriteMarkdownFile(fileName, m.Content)
	log.Printf("MARKDOWN FILE SAVED to disk: %s", fileName)
	// JSON response
	WriteHttpJson(m, w)
}

// Scrapes and returns any valid href (hyperlink)
func LinkHandler(w http.ResponseWriter, r *http.Request) {
	URL := GetURL(r)

	// Instantiate default collector
	c := GetCollector()

	l := &LinkElement{Links: make(map[string]int)}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		log.Printf("A[HREF]::Link request received at %s", e.Request.URL.String())
		log.Printf("LINK(s) EXTRACTED: %s", e.Attr("href"))
		//Tag metadata
		l.Host = e.Request.Host
		l.Method = e.Request.Method
		//Add links
		if link != "" {
			l.Links[link]++
		}
		tagString := strconv.Itoa(l.StatusCode) + "::" + l.Method
		log.Printf("ReqMETA: %s", tagString)
	})

	// extract status code
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
		l.StatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
		l.StatusCode = r.StatusCode
	})

	c.Visit(URL)

	// dump results
	WriteHttpJson(l, w)
}

// Scrapes/dumps all plain text contained in body of the webpage
func TextHandler(w http.ResponseWriter, r *http.Request) {
	URL := GetURL(r)

	// Instantiate default collector
	c := GetCollector()

	p := &PageElement{}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		log.Printf("BODY::Text request received at %s", e.Request.URL.String())
		log.Printf("TEXT EXTRACTED: %s", e.Text)
		//Tag metadata
		p.Host = e.Request.Host
		p.Method = e.Request.Method
		//Add body text
		p.Text = e.Text

		tagString := strconv.Itoa(p.StatusCode) + "::" + p.Method
		log.Printf("ReqMETA: %s", tagString)
	})

	// extract status code
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
		p.StatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
		p.StatusCode = r.StatusCode
	})

	c.Visit(URL)

	// dump results
	WriteHttpJson(p, w)
}

// Scrapes text content from table elements
func TableHandler(w http.ResponseWriter, r *http.Request) {
	URL := GetURL(r)

	// Instantiate default collector
	c := GetCollector()

	var tableTextBuilder strings.Builder

	t := &TableElement{}

	c.OnHTML("tr", func(e *colly.HTMLElement) {
		log.Printf("TR::TEXT request received at %s", e.Request.URL.String())
		log.Printf("TEXT EXTRACTED: %s", e.Text)
		//Tag metadata
		t.Host = e.Request.Host
		t.Method = e.Request.Method
		tableTextBuilder.WriteString(e.Text)

		tagString := strconv.Itoa(t.StatusCode) + "::" + t.Method
		log.Printf("ReqMETA: %s", tagString)
	})

	// extract status code
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
		t.StatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
		t.StatusCode = r.StatusCode
	})

	c.Visit(URL)

	//Add body text
	t.TableText = tableTextBuilder.String()

	// dump results
	WriteHttpJson(t, w)
}
