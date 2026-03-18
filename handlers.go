package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
)

type ScrapeRequest struct {
	StatusCode int
	Host       string
	Method     string
}

type LinkElement struct {
	ScrapeRequest
	Links map[string]int
}

type PageElement struct {
	ScrapeRequest
	Text string
}

//TODO, refactor

func LinkHandler(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	if URL == "" {
		log.Println("missing URL argument")
		return
	}
	log.Println("visiting", URL)

	// Instantiate default collector
	c := colly.NewCollector(
		//colly.AllowedDomains("www.oca.org", "https://news.ycombinator.com/"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./cache"),
		// Cached responses older than the specified duration will be refreshed
		colly.CacheExpiration(24*time.Hour),
	)

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
	b, err := json.Marshal(l)
	if err != nil {
		log.Println("failed to serialize response:", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

func TextHandler(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	if URL == "" {
		log.Println("missing URL argument")
		return
	}
	log.Println("visiting", URL)

	// Instantiate default collector
	c := colly.NewCollector(
		//colly.AllowedDomains("www.oca.org", "https://news.ycombinator.com/"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./cache"),
		// Cached responses older than the specified duration will be refreshed
		colly.CacheExpiration(24*time.Hour),
	)

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
	b, err := json.Marshal(p)
	if err != nil {
		log.Println("failed to serialize response:", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}
