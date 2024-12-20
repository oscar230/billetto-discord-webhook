package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/robfig/cron"
	"gopkg.in/yaml.v3"
)

type Config struct {
	CronExpression string `yaml:"interval"`
	BaseUrl        string `yaml:"base_url"`
	WebhookUrl     string `yaml:"webhook"`
	Event          int    `yaml:"event"`
	Silent         bool   `yaml:"silent"`
	Title          string `yaml:"event_title"`
	Url            string `yaml:"event_url"`
	ImageUrl       string `yaml:"event_image_url"`
}

type EmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

type EmbedThumbnail struct {
	URL string `json:"url,omitempty"`
}

type EmbedImage struct {
	URL string `json:"url,omitempty"`
}

type EmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type Embed struct {
	Author      EmbedAuthor    `json:"author,omitempty"`
	Title       string         `json:"title,omitempty"`
	URL         string         `json:"url,omitempty"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color,omitempty"`
	Fields      []EmbedField   `json:"fields,omitempty"`
	Thumbnail   EmbedThumbnail `json:"thumbnail,omitempty"`
	Image       EmbedImage     `json:"image,omitempty"`
	Footer      EmbedFooter    `json:"footer,omitempty"`
}

type Message struct {
	Username  string  `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
	Content   string  `json:"content,omitempty"`
	Embeds    []Embed `json:"embeds,omitempty"`
}

type StoredAttendees struct {
	Datetime string `json:"datetime"`
	Count    int    `json:"count"`
}

func loadConfig() Config {
	// Open the config file
	file, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal("Error opening config file:", err)
	}
	defer file.Close()

	// Decode the YAML file into the Config struct
	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error decoding config file:", err)
	}
	return config
}

func GetAttendeeCount(config Config) int {
	log.Printf("Getting event %d", config.Event)

	// Get first page
	url := fmt.Sprintf("https://billetto.se/e/%d/attendees?page=1", config.Event)
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
		pattern := fmt.Sprintf(`%d/attendees\?page=`, config.Event)
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
	url = fmt.Sprintf("https://billetto.se/e/%d/attendees?page=%d", config.Event, pageNumberMax)
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

func DiscordSend(currentAttendees, pastAttendees StoredAttendees, config Config) {
	// Create the payload
	var changeText string
	if currentAttendees.Count > pastAttendees.Count {
		changeText = fmt.Sprintf("↗️ Detta är en ökning med %d besökare.\n", currentAttendees.Count-pastAttendees.Count)
	} else if currentAttendees.Count < pastAttendees.Count {
		changeText = fmt.Sprintf("↘️ Detta är en minskning med %d besökare.\n", pastAttendees.Count-currentAttendees.Count)
	} else {
		changeText = ""
	}
	payload := Message{
		Embeds: []Embed{
			{
				Title:       config.Title,
				Description: fmt.Sprintf("# 🎟️ %d besökare", currentAttendees.Count),
				URL:         config.Url,
				Image: EmbedImage{
					URL: config.ImageUrl,
				},
				Fields: []EmbedField{
					{
						Name:   "Föregående mätning",
						Value:  fmt.Sprintf("🗓️ %dst besökare vid mätning %s UTC.", pastAttendees.Count, pastAttendees.Datetime),
						Inline: false,
					},
					{
						Name:   "Förändring",
						Value:  changeText,
						Inline: false,
					},
				},
			},
		},
	}

	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("failed to marshal JSON payload: %w", err)
	}

	// Create the POST request
	req, err := http.NewRequest("POST", config.WebhookUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Fatal("failed to create HTTP request: %w", err)
	}
	log.Printf("HTTP POST %s\n%s", req.URL, req.Body)

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusNoContent {
		log.Fatalf("unexpected response from Discord: %s", resp.Status)
	}
}

func WriteFile(data StoredAttendees) error {
	file, err := os.Create("./event.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON for readability
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func ReadFile() (StoredAttendees, error) {
	var data StoredAttendees
	file, err := os.Open("./event.json")
	if err != nil {
		return data, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return data, fmt.Errorf("failed to decode JSON from file: %w", err)
	}

	return data, nil
}

func Job(config Config) {
	log.Printf("Running job for event %d", config.Event)
	// attendees := GetAttendeeCount(config)
	attendees := rand.Intn(200) // TESTING
	pastAttendees, err := ReadFile()
	if err != nil {
		log.Print("Failed to read past attendees file, will use default values: %w", err)
		pastAttendees = StoredAttendees{
			Datetime: time.Now().UTC().Format(time.RFC3339),
			Count:    -1,
		}
	}
	currentAttendees := StoredAttendees{
		Datetime: time.Now().UTC().Format(time.RFC3339),
		Count:    attendees,
	}
	if currentAttendees.Count != pastAttendees.Count && pastAttendees.Count != -1 {
		DiscordSend(currentAttendees, pastAttendees, config)
	} else {
		log.Printf("Current is %d (now) and past is %d (%s), there is nothing to do.", currentAttendees.Count, pastAttendees.Count, pastAttendees.Datetime)
	}
	WriteFile(currentAttendees)
}

func main() {
	config := loadConfig()
	log.Print("Loaded configuration")

	// Create a new cron scheduler
	scheduler := cron.New()

	// Add a task with a cron expression
	scheduler.AddFunc(config.CronExpression, func() { Job(config) })

	// Start the cron scheduler
	scheduler.Start()
	log.Print("Scheduler started.")

	// Keep the program running to observe scheduled tasks
	select {}
}
