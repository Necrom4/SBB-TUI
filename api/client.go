// Package api
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"sbb-tui/models"
)

func FetchConnections(from string, to string) ([]models.Connection, error) {
	apiURL := fmt.Sprintf("https://transport.opendata.ch/v1/connections?from=%s&to=%s&time=06:00&limit=4",
		url.QueryEscape(from),
		url.QueryEscape(to))

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Connections, nil
}
