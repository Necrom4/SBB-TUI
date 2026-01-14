// Package models
package models

import "time"

type Station struct {
	Name        string  `json:"name"`
	CoordinateX float64 `json:"x"`
	CoordinateY float64 `json:"y"`
}

type Connection struct {
	FromData struct {
		Station   Station   `json:"station"`
		Departure time.Time `json:"departure"`
		Delay     int       `json:"delay"`
		Platform  string    `json:"platform"`
	} `json:"from"`

	ToData struct {
		Station Station   `json:"station"`
		Arrival time.Time `json:"arrival"`
	} `json:"to"`

	Duration  string `json:"duration"`
	Transfers int    `json:"transfers"`
}

type Selection struct {
	From string
	To   string
	Via  string
}

type APIResponse struct {
	Connections []Connection `json:"connections"`
}
