package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- 1. API DATA STRUCTURES ---
// These structs match the JSON returned by transport.opendata.ch

type StationBoardResponse struct {
	Stationboard []Departure `json:"stationboard"`
}

type Departure struct {
	To       string `json:"to"`
	Category string `json:"category"` // e.g., "IC", "S"
	Number   string `json:"number"`   // e.g., "8"
	Stop     struct {
		Departure string `json:"departure"` // ISO timestamp
	} `json:"stop"`
}

// --- 2. MODEL ---
// The application state.

type model struct {
	BernDeparture   *Departure // Stores the next Bern train (nil if loading)
	ZurichDeparture *Departure // Stores the next Zurich train (nil if loading)
	Err             error      // Stores any error that occurred
	width           int        // Terminal width
	height          int        // Terminal height
}

// Initial state of the model
func initialModel() model {
	return model{
		BernDeparture:   nil,
		ZurichDeparture: nil,
	}
}

// --- 3. MESSAGES ---
// Events that happen in our program.

type stationMsg struct {
	stationName string
	data        *Departure
}

type errorMsg error

// --- 4. COMMANDS (API CALLS) ---
// Functions that perform side-effects (like HTTP requests) and return a Message.

func fetchStation(station string) tea.Cmd {
	return func() tea.Msg {
		// 1. Make the HTTP request
		url := fmt.Sprintf("http://transport.opendata.ch/v1/stationboard?station=%s&limit=1", station)
		resp, err := http.Get(url)
		if err != nil {
			return errorMsg(err)
		}
		defer resp.Body.Close()

		// 2. Decode the JSON
		var result StationBoardResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return errorMsg(err)
		}

		// 3. Return the first departure found
		if len(result.Stationboard) > 0 {
			return stationMsg{stationName: station, data: &result.Stationboard[0]}
		}
		return nil
	}
}

// --- 5. BUBBLE TEA FUNCTIONS ---

// Init is run once when the program starts.
func (m model) Init() tea.Cmd {
	// We want to fetch both stations simultaneously
	return tea.Batch(
		fetchStation("Bern"),
		fetchStation("Zurich"),
	)
}

// Update handles incoming messages and modifies the model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}

	// Is it a window resize?
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Is it data arriving from the API?
	case stationMsg:
		if msg.stationName == "Bern" {
			m.BernDeparture = msg.data
		} else if msg.stationName == "Zurich" {
			m.ZurichDeparture = msg.data
		}

	// Is it an error?
	case errorMsg:
		m.Err = msg
	}

	return m, nil
}

// View renders the UI based on the current model.
func (m model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v", m.Err)
	}

	// Define styles using Lipgloss
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")). // Purple-ish border
		Padding(1, 2).                          // Internal padding
		Width(30).                              // Fixed width for the boxes
		Align(lipgloss.Center)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")). // Pink title
		MarginBottom(1)

	// Helper to create the content string for a station
	renderStation := func(name string, dep *Departure) string {
		content := titleStyle.Render(name)
		if dep == nil {
			content += "\nLoading..."
		} else {
			// Parse time to make it look nice (HH:MM)
			t, _ := time.Parse("2006-01-02T15:04:05-0700", dep.Stop.Departure)
			timeStr := t.Format("15:04")

			content += fmt.Sprintf("\nTo: %s\n%s %s\nDep: %s",
				dep.To,
				dep.Category,
				dep.Number,
				timeStr,
			)
		}
		return boxStyle.Render(content)
	}

	// Render the two boxes
	bernBox := renderStation("Bern", m.BernDeparture)
	zurichBox := renderStation("Zurich", m.ZurichDeparture)

	// Join them horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, bernBox, zurichBox)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
