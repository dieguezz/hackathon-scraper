package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gocolly/colly/v2"
)

// ScrapeHackathon srapes Hackathon.com hackathons
func ScrapeHackathon() {
	log.Println("Scraping Hackathon.com")
	fName := "hackathon.json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	events := make([]Event, 0, 200)

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("hackathon.com", "www.hackathon.com"),
		colly.URLFilters(regexp.MustCompile("online|event")),
		colly.CacheDir("./hackathon_list_cache"),
		colly.Async(),
	)

	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnError(func(e *colly.Response, err error) {
		log.Fatal(err)
		return
	})

	// For every page, visit its link
	c.OnHTML(`[data-pagination-more-link="hackathons"]`, func(e *colly.HTMLElement) {
		link := e.Attr("href")
		c.Visit(link)
	})

	// For every item in page, visit its link
	c.OnHTML(`.ht-eb-card__title`, func(e *colly.HTMLElement) {
		link := e.Attr("href")
		c.Visit(link)
	})

	type ld struct {
		Name        string
		Description string
		StartDate   string
		EndDate     string
		URL         string
	}

	c.OnHTML(`head`, func(e *colly.HTMLElement) {

		var data []ld

		ldJSON := e.ChildText(`[type="application/ld+json"]`)

		json.Unmarshal([]byte(ldJSON), &data)

		for i := range data {
			item := data[i]
			start, _ := time.Parse(time.RFC3339, item.StartDate)
			end, _ := time.Parse(time.RFC3339, item.EndDate)

			events = append(events, Event{
				item.Name,
				item.Description,
				start,
				end,
				item.URL,
			})
		}
	})

	// Start scraping on
	c.Visit("http://www.hackathon.com/online")
	c.Wait()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(events)
}
