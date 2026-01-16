// Package api
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"sbb-tui/models"
	"sbb-tui/utils"
)

func FetchConnections(from, to, date, timeStr string, isArrivalTime bool) ([]models.Connection, error) {
	now := time.Now()

	if date == "" {
		date = now.Format("2006-01-02")
	}

	if timeStr == "" {
		timeStr = now.Format("15:04")
	}

	apiURL := fmt.Sprintf(
		"https://transport.opendata.ch/v1/connections?from=%s&to=%s&date=%s&time=%s&isArrivalTime=%s&limit=6",
		url.QueryEscape(from),
		url.QueryEscape(to),
		url.QueryEscape(date),
		url.QueryEscape(timeStr),
		strconv.Itoa(utils.Btoi(isArrivalTime)),
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
