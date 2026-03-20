package main

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

type TableElement struct {
	ScrapeRequest
	TableText string
}
