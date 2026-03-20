package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
)

// Return files for Logging or dumping
func GetFileWrite(fileName string) *os.File {
	if fileName == "" {
		log.Fatalf("errors.New(\"\"): %v\n", errors.New("WRITE FILE ERROR"))
		return nil
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("errors.New(\"\"): %v\n", err)
		return nil
	}

	return file
}

func GetCollector() *colly.Collector {
	// Instantiate default collector
	return colly.NewCollector(
		//colly.AllowedDomains("www.oca.org", "https://news.ycombinator.com/"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./cache"),
		// Cached responses older than the specified duration will be refreshed
		colly.CacheExpiration(24*time.Hour),
	)

}

func GetURL(r *http.Request) string {
	URL := r.URL.Query().Get("url")
	if URL == "" {
		missingMsg := "missing URL argument"
		log.Println(missingMsg)
		return errors.New(missingMsg).Error()
	}

	log.Println("visiting", URL)

	return URL
}

func WriteHttpJson(v any, w http.ResponseWriter) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Println("failed to serialize response:", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}
