package main

//Link Server (microservice)
//Get links from any website, with associated metadata, to fuel faster downstream scraping from other programs

//See TODO.txt

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
)

type pageInfo struct {
	StatusCode int
	Links      map[string]int
}

func handler(w http.ResponseWriter, r *http.Request) {
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

	p := &pageInfo{Links: make(map[string]int)}

	// count links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			p.Links[link]++
		}
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

func main() {
	// example usage: curl -s 'http://127.0.0.1:7171/?url=http://go-colly.org/'
	addr := ":7171"

	http.HandleFunc("/", handler)

	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
