package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DatabaseUrl    string `yaml:"database"`
	CronExpression string `yaml:"interval"`
	Delay          int    `yaml:"delay"`
	BaseUrl        string `yaml:"base_url"`
	UserAgent      string `yaml:"user_agent"`
	WebhookUrl     string `yaml:"webhook"`
	Event          int    `yaml:"event"`
	Silent         bool   `yaml:"silent"`
}

type DiscordWebhookPayload struct {
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
	Embeds    []struct {
		Author struct {
			Name    string `json:"name,omitempty"`
			URL     string `json:"url,omitempty"`
			IconURL string `json:"icon_url,omitempty"`
		} `json:"author,omitempty"`
		Title       string `json:"title,omitempty"`
		URL         string `json:"url,omitempty"`
		Description string `json:"description,omitempty"`
		Color       int    `json:"color,omitempty"`
		Fields      []struct {
			Name   string `json:"name,omitempty"`
			Value  string `json:"value,omitempty"`
			Inline bool   `json:"inline,omitempty"`
		} `json:"fields,omitempty"`
		Thumbnail struct {
			URL string `json:"url,omitempty"`
		} `json:"thumbnail,omitempty"`
		Image struct {
			URL string `json:"url,omitempty"`
		} `json:"image,omitempty"`
		Footer struct {
			Text    string `json:"text,omitempty"`
			IconURL string `json:"icon_url,omitempty"`
		} `json:"footer,omitempty"`
	} `json:"embeds,omitempty"`
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

func DiscordSend(webhookUrl string, attendeesCount int) {
	// Create the payload
	payload := DiscordWebhookPayload{
		Content: fmt.Sprintf("Test %d", attendeesCount),
	}

	// Convert the payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("failed to marshal JSON payload: %w", err)
	}

	// Create the POST request
	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(payloadBytes))
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

	log.Printf("Webhook sent.")
}

func main() {
	config := loadConfig()
	log.Print("Loaded configuration")

	attendees := GetAttendeeCount(config)
	DiscordSend(config.WebhookUrl, attendees)
	os.Exit(1)

	// Create a new cron scheduler
	scheduler := cron.New()

	// Add a task with a cron expression
	scheduler.AddFunc(config.CronExpression, func() {
		attendees := GetAttendeeCount(config)
		DiscordSend(config.WebhookUrl, attendees)
	})

	// Start the cron scheduler
	scheduler.Start()
	log.Print("Scheduler started.")

	// Keep the program running to observe scheduled tasks
	select {}
}
