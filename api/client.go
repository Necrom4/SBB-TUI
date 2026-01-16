// Package api
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"sbb-tui/models"
)

func FetchConnections(from, to, date, timeStr string) ([]models.Connection, error) {
	now := time.Now()

	if date == "" {
		date = now.Format("2006-01-02")
	}

	if timeStr == "" {
		timeStr = now.Format("15:04")
	}

	apiURL := fmt.Sprintf(
		"https://transport.opendata.ch/v1/connections?from=%s&to=%s&date=%s&time=%s&limit=6",
		url.QueryEscape(from),
		url.QueryEscape(to),
		url.QueryEscape(date),
		url.QueryEscape(timeStr),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Connections, nil
}
