package billetto

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Event struct {
	ID            string    `json:"id,omitempty"`
	Object        string    `json:"object,omitempty"`
	Name          string    `json:"name,omitempty"`
	Currency      string    `json:"currency,omitempty"`
	State         string    `json:"state,omitempty"`
	Public        bool      `json:"public,omitempty"`
	OnlineEvent   bool      `json:"online_event,omitempty"`
	Kind          string    `json:"kind,omitempty"`
	TotalCapacity int       `json:"total_capacity,omitempty"`
	Parent        any       `json:"parent,omitempty"`
	PublicURL     string    `json:"public_url,omitempty"`
	StartsAt      time.Time `json:"starts_at,omitempty"`
	EndsAt        time.Time `json:"ends_at,omitempty"`
	PublishedAt   time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
	Availability  struct {
		Available int    `json:"available,omitempty"`
		Status    string `json:"status,omitempty"`
	} `json:"availability,omitempty"`
	Venue        any    `json:"venue,omitempty"`
	Organization string `json:"organization,omitempty"`
	Location     string `json:"location,omitempty"`
	Plans        struct {
		Object  string `json:"object,omitempty"`
		Data    []any  `json:"data,omitempty"`
		HasMore bool   `json:"has_more,omitempty"`
		Total   int    `json:"total,omitempty"`
	} `json:"plans,omitempty"`
}

type EventAttendees struct {
	Object string `json:"object,omitempty"`
	Data   []struct {
		ID           string    `json:"id,omitempty"`
		Object       string    `json:"object,omitempty"`
		Barcode      string    `json:"barcode,omitempty"`
		State        string    `json:"state,omitempty"`
		FeeIncluded  bool      `json:"fee_included,omitempty"`
		Name         string    `json:"name,omitempty"`
		Email        string    `json:"email,omitempty"`
		Photo        any       `json:"photo,omitempty"`
		Type         string    `json:"type,omitempty"`
		TicketBuyer  string    `json:"ticket_buyer,omitempty"`
		Order        string    `json:"order,omitempty"`
		Price        int       `json:"price,omitempty"`
		Fee          int       `json:"fee,omitempty"`
		CreatedAt    time.Time `json:"created_at,omitempty"`
		UpdatedAt    time.Time `json:"updated_at,omitempty"`
		Space        any       `json:"space,omitempty"`
		AddressLine1 string    `json:"address_line_1,omitempty"`
		AddressLine2 string    `json:"address_line_2,omitempty"`
		PostalCode   string    `json:"postal_code,omitempty"`
		City         string    `json:"city,omitempty"`
		CountryCode  string    `json:"country_code,omitempty"`
		PhoneNumber  string    `json:"phone_number,omitempty"`
		Scannings    struct {
			Object  string `json:"object,omitempty"`
			Data    []any  `json:"data,omitempty"`
			HasMore bool   `json:"has_more,omitempty"`
			Total   int    `json:"total,omitempty"`
		} `json:"scannings,omitempty"`
		BookingQuestionResponses struct {
			Object  string `json:"object,omitempty"`
			Data    []any  `json:"data,omitempty"`
			HasMore bool   `json:"has_more,omitempty"`
			Total   int    `json:"total,omitempty"`
		} `json:"booking_question_responses,omitempty"`
		OrderLine            string `json:"order_line,omitempty"`
		TicketType           string `json:"ticket_type,omitempty"`
		Event                string `json:"event,omitempty"`
		Membership           any    `json:"membership,omitempty"`
		Subscription         any    `json:"subscription,omitempty"`
		NewsletterPermission bool   `json:"newsletter_permission,omitempty"`
	} `json:"data,omitempty"`
	HasMore bool   `json:"has_more,omitempty"`
	Total   int    `json:"total,omitempty"`
	URL     string `json:"url,omitempty"`
	NextURL string `json:"next_url,omitempty"`
}

func GetEventInfo(eventId int, AccessKeyId, AccessKeySecret string) Event {
	// Build url
	url := fmt.Sprintf("https://billetto.dk/api/v3/organiser/events/%d", eventId)

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Keypair", fmt.Sprintf("%s:%s", AccessKeyId, AccessKeySecret))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	// Unmarshal JSON into struct
	var data Event
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	return data
}

func GetEventAttendees(eventId int, AccessKeyId, AccessKeySecret string) EventAttendees {
	// Build url
	url := fmt.Sprintf("https://billetto.dk/api/v3/organiser/events/%d/attendees?limit=0", eventId)

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Keypair", fmt.Sprintf("%s:%s", AccessKeyId, AccessKeySecret))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response body:", err)
	}

	// Unmarshal JSON into struct
	var data EventAttendees
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	return data
}
