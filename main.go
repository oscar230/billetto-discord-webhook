package main

import (
	"fmt"
	"log"
	"oscar230/billetto-discord-webhook/discord"
	"time"
)

func Job(webhookUrl string, eventId int, eventTitle, eventUrl, eventImageUrl, priceCurrency string, priceList []Price) {
	log.Printf("Running job for event %d", eventId)

	// Get current attendees
	eventAttendeeCount := 1233 // billetto.EventAttendeeCount(eventId)
	currentAttendees := Attendees{
		Datetime: time.Now().UTC().Format(time.RFC3339),
		Count:    eventAttendeeCount,
	}

	// Get past attendees
	pastAttendees, err := Load()
	if err != nil {
		log.Print("Failed to read past attendees file, will use default values: %w", err)
		pastAttendees = Attendees{
			Datetime: time.Now().UTC().Format(time.RFC3339),
			Count:    -1,
		}
	}

	// If attendees have changed since last mesaurement, send a message otherwise ignore
	if currentAttendees.Count != pastAttendees.Count && pastAttendees.Count != -1 {
		// Create the message variables
		var changeText string
		if currentAttendees.Count >= pastAttendees.Count {
			changeText = fmt.Sprintf("↗️ Detta är en ökning med %d besökare.\n", currentAttendees.Count-pastAttendees.Count)
		} else {
			changeText = fmt.Sprintf("↘️ Detta är en minskning med %d besökare.\n", pastAttendees.Count-currentAttendees.Count)
		}

		// Create the message
		inlineFields := false
		message := discord.Message{
			Embeds: []discord.Embed{
				{
					Title:       eventTitle,
					Description: fmt.Sprintf("# 🎟️ %d besökare", currentAttendees.Count),
					URL:         eventUrl,
					Image: discord.EmbedImage{
						URL: eventImageUrl,
					},
					Fields: []discord.EmbedField{
						{
							Name:   "💸 Intäkt",
							Value:  GetRevenue(priceList, priceCurrency, currentAttendees.Count),
							Inline: inlineFields,
						},
						{
							Name:   "Förändring",
							Value:  changeText,
							Inline: inlineFields,
						},
						{
							Name:   "Föregående mätning",
							Value:  fmt.Sprintf("🗓️ %dst besökare vid mätning %s UTC.", pastAttendees.Count, pastAttendees.Datetime),
							Inline: inlineFields,
						},
					},
				},
			},
		}
		discord.Send(webhookUrl, message)
	} else {
		log.Printf("Current is %d (now) and past is %d (%s), there is nothing to do.", currentAttendees.Count, pastAttendees.Count, pastAttendees.Datetime)
	}
	Store(currentAttendees)
}

func main() {
	config := loadConfig()
	log.Print("Loaded configuration")

	Job(config.WebhookUrl, config.Event, config.Title, config.Url, config.ImageUrl, config.PriceCurrency, config.PriceList)

	// // Create a new cron scheduler
	// scheduler := cron.New()

	// // Add a task with a cron expression
	// scheduler.AddFunc(config.CronExpression, func() { Job(config.WebhookUrl, config.Event, config.Title, config.Url, config.ImageUrl) })

	// // Start the cron scheduler
	// scheduler.Start()
	// log.Printf("Scheduler started at system time: %s", time.Now())

	// // Keep the program running to observe scheduled tasks
	// select {}
}
