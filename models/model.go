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
	CoordinateX float64 `json:"x"`
	CoordinateY float64 `json:"y"`
}

type Connection struct {
	FromData struct {
		Station   Station       `json:"station"`
		Departure SBBDateLayout `json:"departure"`
		Delay     int           `json:"delay"`
		Platform  string        `json:"platform"`
	} `json:"from"`

	ToData struct {
		Station Station       `json:"station"`
		Arrival SBBDateLayout `json:"arrival"`
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
