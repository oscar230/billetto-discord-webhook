package main

import (
	"fmt"
	"log"
	"oscar230/billetto-discord-webhook/billetto"
	"oscar230/billetto-discord-webhook/discord"
	"time"

	"github.com/robfig/cron/v3"
)

func Job(webhookUrl string, eventId int, eventTitle, eventUrl, eventImageUrl, priceCurrency string, priceList []Price) {
	log.Printf("Running job for event %d", eventId)

	// Get current attendees
	eventAttendeeCount := billetto.EventAttendeeCount(eventId)
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
			changeText = fmt.Sprintf("Detta är en ökning med %d besökare.\n", currentAttendees.Count-pastAttendees.Count)
		} else {
			changeText = fmt.Sprintf("Detta är en minskning med %d besökare.\n", pastAttendees.Count-currentAttendees.Count)
		}

		// Create the message
		inlineFields := false
		message := discord.Message{
			Embeds: []discord.Embed{
				{
					Title:       eventTitle,
					Description: fmt.Sprintf("# %d st sålda biljetter\n*Intäkternas summa baseras på antagandet att alla gäster köper den billigaste biljetten tillgänglig. Intäkter är inkl. moms. Detta program kollar Billetto varje dag 12:00, om ingen förändring har skett så skickas inget meddelande.*\n", currentAttendees.Count),
					URL:         eventUrl,
					Image: discord.EmbedImage{
						URL: eventImageUrl,
					},
					Footer: discord.EmbedFooter{
						Text: "Av Oscar, för Klanglandet.",
					},
					Fields: []discord.EmbedField{
						{
							Name:   "Intäkter",
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
							Value:  fmt.Sprintf("%dst besökare vid mätning %s UTC.", pastAttendees.Count, pastAttendees.Datetime),
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

	// Create a new cron scheduler
	scheduler := cron.New()

	// Add a task with a cron expression
	scheduler.AddFunc(config.CronExpression, func() {
		Job(config.WebhookUrl, config.Event, config.Title, config.Url, config.ImageUrl, config.PriceCurrency, config.PriceList)
	})

	// Start the cron scheduler
	scheduler.Start()
	log.Printf("Scheduler started at system time: %s", time.Now())

	// Keep the program running to observe scheduled tasks
	select {}
}
