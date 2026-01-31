// Package api
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sbb-tui/models"
	"sbb-tui/utils"
)

func FetchConnections(from, to, date, timeStr string, isArrivalTime bool, limit int) ([]models.Connection, error) {
	parts := []string{
		fmt.Sprintf("from=%s", url.QueryEscape(from)),
		fmt.Sprintf("to=%s", url.QueryEscape(to)),
	}

	if date != "" {
		parts = append(parts, fmt.Sprintf("date=%s", url.QueryEscape(date)))
	}

	if timeStr != "" {
		parts = append(parts, fmt.Sprintf("time=%s", url.QueryEscape(timeStr)))
	}

	parts = append(parts,
		fmt.Sprintf("isArrivalTime=%s", strconv.Itoa(utils.Btoi(isArrivalTime))),
		fmt.Sprintf("limit=%v", limit),
	)

	apiURL := "https://transport.opendata.ch/v1/connections?" + strings.Join(parts, "&")

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
