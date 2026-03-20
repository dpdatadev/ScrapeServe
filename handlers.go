package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

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
