// Package models
package models

import (
	"strings"
	"time"
)

type SBBDateLayout struct {
	time.Time
}

func (st SBBDateLayout) Sub(other SBBDateLayout) time.Duration {
	return st.Time.Sub(other.Time)
}

func (st *SBBDateLayout) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02T15:04:05-0700", s)
	if err != nil {
		return err
	}
	st.Time = t
	return nil
}

type Departure struct {
	Departure SBBDateLayout `json:"departure"`
	Delay     int           `json:"delay"`
}

type Arrival struct {
	Arrival SBBDateLayout `json:"arrival"`
	Delay   int           `json:"delay"`
}

type Station struct {
	Name        string  `json:"name"`
	CoordinateX float64 `json:"coordinate.x"`
	CoordinateY float64 `json:"coordinate.y"`
}

type Section struct {
	Journey *struct {
		Category string `json:"category"`
		Number   string `json:"number"`
		Operator string `json:"operator"`
		To       string `json:"to"`
	} `json:"journey"`
	Walk *struct {
		Duration  int       `json:"duration"`
		Departure Departure `json:"departure"`
		Arrival   Arrival   `json:"arrival"`
	} `json:"walk"`
	Departure Departure `json:"departure"`
	Arrival   Arrival   `json:"arrival"`
}

type Connection struct {
	FromData struct {
		Station   Station       `json:"station"`
		Departure SBBDateLayout `json:"departure"`
		Delay     int           `json:"delay"`
		Platform  string        `json:"platform"`
	} `json:"from"`

	ToData struct {
		Station  Station       `json:"station"`
		Arrival  SBBDateLayout `json:"arrival"`
		Platform string        `json:"platform"`
	} `json:"to"`

	Duration    string    `json:"duration"`
	Transfers   int       `json:"transfers"`
	Capacity1st string    `json:"capacity1st"`
	Capacity2nd string    `json:"capacity2nd"`
	Sections    []Section `json:"sections"`
}

type Input struct {
	From            string
	To              string
	Via             string
	Date            time.Time
	Time            time.Time
	IsArrivalTime   bool
	Transportations string
	Limit           int
}

type APIResponse struct {
	Connections []Connection `json:"connections"`
}
