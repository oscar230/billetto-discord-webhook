package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CronExpression string `yaml:"interval"`
	WebhookUrl     string `yaml:"webhook"`
	Event          int    `yaml:"event"`
	Title          string `yaml:"event_title"`
	Url            string `yaml:"event_url"`
	ImageUrl       string `yaml:"event_image_url"`
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
