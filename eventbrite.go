package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
)

// ScrapeEventBrite srapes Eventbrite hackathons
func ScrapeEventBrite() {
	log.Println("Scraping Eventbrite.com")
	fName := "eventbrite.json"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("eventbrite.com", "www.eventbrite.com"),
		colly.URLFilters(regexp.MustCompile("online/hackathon")),
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./eventbrite_list_cache"),
	)

	// Create another collector to scrape event details
	detailCollector := colly.NewCollector(
		colly.AllowedDomains("eventbrite.com", "www.eventbrite.com", "www.eventbrite.co.uk", "eventbrite.co.uk"),
		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./eventbrite_detail_cache"),
		colly.Async(),
	)

	events := make([]Event, 0, 200)

	// For every event card, visit its link
	c.OnHTML(`.eds-event-card-content__primary-content > a[tabindex="0"]`, func(e *colly.HTMLElement) {
		link := e.Attr("href")
		detailCollector.Visit(link)
	})

	// For every event card, visit its link
	c.OnHTML(`[data-spec="page-next"]`, func(e *colly.HTMLElement) {
		page, err := strconv.Atoi(e.Request.URL.Query().Get("page"))
		if err != nil {
			log.Fatal("Page error")
			return
		}
		page++
		baseURL := "https://www.eventbrite.com/d/online/hackathon/?page="
		log.Println("PAGE", fmt.Sprint(baseURL, page))

		c.Visit(fmt.Sprint(baseURL, page))
	})

	// Extract details of the eventº
	detailCollector.OnHTML("html", func(e *colly.HTMLElement) {
		dateAttrs := e.ChildAttrs(".g-cell >.event-details__data meta", "content")
		start := time.Now()
		end := time.Now()
		if len(dateAttrs) > 0 {
			start, _ = time.Parse(time.RFC3339, dateAttrs[0])
			end, _ = time.Parse(time.RFC3339, dateAttrs[1])
		}
		event := Event{
			e.ChildAttr(`meta[name="twitter:title"]`, "content"),
			e.ChildText(`[data-automation="listing-event-description"]`),
			start,
			end,
			e.Request.URL.String(),
		}
		events = append(events, event)
	})

	detailCollector.Wait()

	// Start scraping on
	c.Visit("https://www.eventbrite.com/d/online/hackathon/?page=1")

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	// Dump json to the standard output
	enc.Encode(events)
}
