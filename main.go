package main

import (
	"fmt"
	"sync"
	"time"
)

// Event stores information about a eventbrite event
type Event struct {
	Title       string
	Description string
	Start       time.Time
	End         time.Time
	URL         string
}

func main() {
	var wg sync.WaitGroup
	scrapers := [...]func(){ScrapeEventBrite, ScrapeHackathon}
	wg.Add(len(scrapers))

	for _, scrape := range scrapers {
		go func(scrape func()) {
			scrape()
			wg.Done()
		}(scrape)
	}

	wg.Wait()
	fmt.Println("done")
}
