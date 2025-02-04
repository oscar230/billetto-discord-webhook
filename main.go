package main

import (
	"fmt"
	"log"
	"oscar230/billetto-discord-webhook/billetto"
	"oscar230/billetto-discord-webhook/discord"
	"time"

	"github.com/robfig/cron/v3"
)

func Job(config Config) {
	log.Printf("Running job for event %d", config.BillettoEventId)

	// Get current attendees
	eventAttendees, xRatelimitRemaining := billetto.GetEventAttendees(config.BillettoEventId, config.BillettoAccessKeyId, config.BillettoAccessKeySecret)
	log.Printf("Billetto API points left: %s", xRatelimitRemaining)
	currentAttendees := Attendees{
		Datetime: time.Now().UTC().Format(time.RFC3339),
		Count:    eventAttendees.Total,
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
			changeText = fmt.Sprintf("Detta är en ökning med %d besökare sedan senaste mätning.\n", currentAttendees.Count-pastAttendees.Count)
		} else {
			changeText = fmt.Sprintf("Detta är en minskning med %d besökare sedan senaste mätning..\n", pastAttendees.Count-currentAttendees.Count)
		}

		// Create the message
		inlineFields := false
		message := discord.Message{
			Embeds: []discord.Embed{
				{
					Title:       config.BillettoEventName,
					Description: fmt.Sprintf("# %d st sålda biljetter", currentAttendees.Count),
					URL:         config.BillettoEventUrl,
					Image: discord.EmbedImage{
						URL: config.BillettoEventImageUrl,
					},
					Fields: []discord.EmbedField{
						{
							Name:   "Förändring",
							Value:  changeText,
							Inline: inlineFields,
						},
						{
							Name:   "Föregående mätning",
							Value:  fmt.Sprintf("%dst besökare vid mätning %s UTC.", pastAttendees.Count, pastAttendees.Datetime),
							Inline: inlineFields,
						},
					},
				},
			},
		}
		discord.Send(config.DiscordWebhookUrl, message)
	} else {
		log.Printf("Current is %d (now) and past is %d (%s), there is nothing to do.", currentAttendees.Count, pastAttendees.Count, pastAttendees.Datetime)
	}
	Store(currentAttendees)
}

func main() {
	config := loadConfig()
	log.Print("Loaded configuration")

	// Create a new cron scheduler
	scheduler := cron.New()

	// Add a task with a cron expression
	scheduler.AddFunc(config.SchedulerCronExpression, func() {
		Job(config)
	})

	// Start the cron scheduler
	scheduler.Start()
	log.Printf("Scheduler started at system time: %s", time.Now())

	// Keep the program running to observe scheduled tasks
	select {}
}
