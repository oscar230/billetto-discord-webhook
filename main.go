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
	UserAgent      string     `yaml:"user_agent"`
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

func Scrape(config Config) Event {
	// Parse the URL
	baseUrl, err := url.Parse(config.BaseUrl)
	if err != nil {
		log.Fatal("Error parsing URL: ", err)
	}

	// if baseUrl.Port() == "" {
	// 	log.Fatal("Error parsing URL, port is not set: ", baseUrl)
	// }

	// // Parse host from URL
	// host, _, err := net.SplitHostPort(baseUrl.Host)
	// if err != nil {
	// 	log.Fatal("Error parsing URL host: ", err)
	// }

	// Validate user agent
	if len(config.UserAgent) < 8 {
		log.Fatal("User agent too small, make sure it contains contact information: ", config.UserAgent)
	}

	// Create collector
	collector := colly.NewCollector(
		colly.AllowedDomains(baseUrl.Hostname()),
		colly.AllowURLRevisit(),
		colly.UserAgent(config.UserAgent),
	)

	// Set error handler
	collector.OnError(func(r *colly.Response, err error) {
		log.Print("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Set limiter
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       3 * time.Second,
		Parallelism: 2,
	})

	// Before making a request
	collector.OnRequest(func(r *colly.Request) {
		log.Print("Visiting: ", r.URL.String())
	})

	// On every a element which has href attribute
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		log.Printf("Link found: %q -> %s\n", e.Text, link)
	})

	// Visit event's page
	visitUrl := baseUrl.JoinPath("e/" + config.Webhooks[0].Events[0]).String()
	err = collector.Visit(visitUrl)
	if err != nil {
		log.Fatal("Error visiting: ", err)
	}

	return Event{}
}

func main() {
	config := loadConfig()
	log.Print("Loaded configuration")

	Scrape(config)
	log.Fatal("Done")

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
