package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Attendees struct {
	Datetime string `json:"datetime"`
	Count    int    `json:"count"`
}

func Store(data Attendees) error {
	file, err := os.Create("./event.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON for readability
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func Load() (Attendees, error) {
	var data Attendees
	file, err := os.Open("./event.json")
	if err != nil {
		return data, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return data, fmt.Errorf("failed to decode JSON from file: %w", err)
	}

	return data, nil
}
