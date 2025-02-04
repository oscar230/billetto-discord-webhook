package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SchedulerCronExpression string `yaml:"cron"`
	DiscordWebhookUrl       string `yaml:"discord_webhook"`
	BillettoEventId         int    `yaml:"billetto_event_id"`
	BillettoEventName       string `yaml:"billetto_event_name"`
	BillettoEventUrl        string `yaml:"billetto_event_url"`
	BillettoEventImageUrl   string `yaml:"billetto_event_image_url"`
	BillettoAccessKeyId     string `yaml:"billetto_access_key_id"`
	BillettoAccessKeySecret string `yaml:"billetto_access_key_secret"`
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
