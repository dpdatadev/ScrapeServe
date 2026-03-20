package main

//Scrape Server (microservice)
//Get links from any website, with associated metadata, to fuel faster downstream scraping from other programs

//See TODO.txt

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var sessionID string

func init() {
	log.SetPrefix("[=]SCRAPE SERVE[=]")
	log.SetFlags(0)
	localLog := GetFileWrite("executions.log")
	log.SetOutput(localLog)
	log.Printf("SCRAPE BEGIN: %s\n", time.Now().Local().String())
}

func main() {
	// example usage: curl -s 'http://127.0.0.1:7171/links?url=http://go-colly.org/'
	// example usage: curl -s 'http://127.0.0.1:7171/text?url=http://go-colly.org/'
	addr := ":7171"
	sessionid := uuid.NewString()
	welcomeMessage := fmt.Sprintf("::SCRAPE SERVE:: => listening for [/links] and [/text] on %s", addr)

	http.HandleFunc("/links", LinkHandler)
	http.HandleFunc("/text", TextHandler)
	http.HandleFunc("/table", TableHandler)
	log.Printf("Starting NEW SESSION ( %s )- CACHE, enabled (YES)::Duration(24 Hours)\n", sessionid)
	log.Println(welcomeMessage) //logfile
	fmt.Println(welcomeMessage) //console
	log.Fatal(http.ListenAndServe(addr, nil))
}
