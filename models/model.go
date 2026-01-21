// Package models
package models

import (
	"strings"
	"time"
)

const sbbDateLayout = "2006-01-02T15:04:05-0700"

type SBBDateLayout struct {
	time.Time
}

func (st *SBBDateLayout) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		return nil
	}
	t, err := time.Parse(sbbDateLayout, s)
	if err != nil {
		return err
	}
	st.Time = t
	return nil
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
		Duration int `json:"duration"`
	} `json:"walk"`
	Departure struct {
		Departure SBBDateLayout `json:"departure"`
		Delay     int           `json:"delay"`
	} `json:"departure"`
	Arrival struct {
		Arrival SBBDateLayout `json:"arrival"`
		Delay   int           `json:"delay"`
	} `json:"arrival"`
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
