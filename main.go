package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gocolly/colly"
	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"
)

type Webhooks struct {
	WebhookUrl string   `yaml:"url"`
	Events     []string `yaml:"events"`
	Enabled    bool     `yaml:"enabled"`
	Silent     bool     `yaml:"silent"`
}

type Config struct {
	DatabaseUrl    string     `yaml:"database"`
	CronExpression string     `yaml:"interval"`
	Webhooks       []Webhooks `yaml:"webhooks"`
	BaseUrl        string     `yaml:"base_url"`
}

type EventAttendee struct {
	Url      string
	Name     string
	ImageUrl string
}

type EventOrganizer struct {
	Url      string
	Name     string
	ImageUrl string
}

type Event struct {
	Id        string
	Name      string
	ImageUrl  string
	Organizer EventOrganizer
	Attendees []EventAttendee
}

func loadConfig() Config {
	// Open the config file
	file, err := os.Open("config.yaml")
	if err != nil {
		log.Fatal("Error opening config file:", err)
		os.Exit(1)
	}
	defer file.Close()

	// Decode the YAML file into the Config struct
	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error decoding config file:", err)
		os.Exit(1)
	}
	return config
}

func Scrape(config Config) Event {
	// Parse the URL
	baseUrl, err := url.Parse(config.BaseUrl)
	if err != nil {
		log.Fatal("Error parsing URL:", err)
		os.Exit(1)
	}

	collector := colly.NewCollector(
		colly.AllowedDomains(baseUrl.Host),
		colly.AllowURLRevisit(),
		colly.Async(true),
	)

	// Set error handler
	collector.OnError(func(r *colly.Response, err error) {
		log.Print("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	collector.Limit(&colly.LimitRule{
		DomainGlob:  baseUrl.Host + "/*",
		RandomDelay: 2 * time.Second,
		Delay:       3 * time.Second,
	})

	// Before making a request
	collector.OnRequest(func(r *colly.Request) {
		log.Print("Visiting", r.URL.String())
	})

	// On every a element which has href attribute
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Printf("Link found: %q -> %s\n", e.Text, link)
	})

	err = collector.Visit("https://billetto.se/e/1064241/attendees")
	if err != nil {
		log.Fatal("Error visiting: ", err)
		os.Exit(1)
	}

	return Event{}
}

func main() {
	config := loadConfig()
	log.Print("Loaded configuration")

	Scrape(config)
	log.Print("Done")
	os.Exit(0)

	// Create a new cron scheduler
	scheduler := cron.New()

	// Add a task with a cron expression
	scheduler.AddFunc(config.CronExpression, func() { Scrape(config) })

	// Start the cron scheduler
	scheduler.Start()
	log.Print("Scheduler started.")

	// Keep the program running to observe scheduled tasks
	select {}
}
