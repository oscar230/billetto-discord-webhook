package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

func main() {
	config := loadConfig()
	log.Print("Loaded Configuration.")

	// Create a new cron scheduler
	scheduler := cron.New()

	// Add a task with a cron expression
	scheduler.AddFunc(config.CronExpression, func() {
		fmt.Println("Task running at", time.Now())
	})

	// Start the cron scheduler
	scheduler.Start()
	log.Print("Scheduler started.")

	// Keep the program running to observe scheduled tasks
	select {}
}
