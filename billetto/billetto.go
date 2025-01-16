package billetto

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func EventAttendeeCount(eventId int) int {
	log.Printf("Getting event %d", eventId)

	// Get first page
	url := fmt.Sprintf("https://billetto.se/e/%d/attendees?page=1", eventId)
	log.Printf("HTTP GET %s", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to fetch the page: %w", err)
	}
	defer resp.Body.Close()

	// Check http response status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("non-OK HTTP status: %s", resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Failed to parse HTML: %w", err)
	}

	// Extract matching links
	var pages []int
	doc.Find("a").Each(func(index int, element *goquery.Selection) {
		// Pattern to match links
		pattern := fmt.Sprintf(`%d/attendees\?page=`, eventId)
		// Define a regex to match the pattern
		re := regexp.MustCompile(pattern)
		href, exists := element.Attr("href")
		if exists && re.MatchString(href) {
			pageNumber, err := strconv.Atoi(strings.Split(href, "?page=")[1])
			if err != nil {
				log.Fatal("Failed to conert page number to int: %w", err)
			}
			pages = append(pages, pageNumber)
			// log.Printf("Page: %d", pageNumber)
		}
	})

	// Get the maximum page value
	pageNumberMax := slices.Max(pages)
	log.Printf("Last page: %d", pageNumberMax)

	// Get last page
	url = fmt.Sprintf("https://billetto.se/e/%d/attendees?page=%d", eventId, pageNumberMax)
	log.Printf("HTTP GET %s", url)
	resp, err = http.Get(url)
	if err != nil {
		log.Fatal("Failed to fetch the page: %w", err)
	}
	defer resp.Body.Close()

	// Check http response status code
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("non-OK HTTP status: %s", resp.Status)
	}

	// Load the HTML document
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Failed to parse HTML: %w", err)
	}

	countAttendees := doc.Find("#main-content > main > div > div.grid > div.relative").Length()
	log.Printf("Last page attendees count: %d", countAttendees)
	countAttendees += (pageNumberMax - 1) * 4 * 3
	log.Printf("Total attendees count: %d", countAttendees)

	return countAttendees
}
