package webhook

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

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

func Send(webhookUrl string, message Message) {
	// Convert the payload to JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Fatal("failed to marshal JSON message: %w", err)
	}

	// Create the POST request
	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(messageBytes))
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
