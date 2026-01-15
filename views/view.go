// Package views
package views

import (
	"fmt"
	"strings"

	"sbb-tui/api"
	"sbb-tui/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	noStyle        = lipgloss.NewStyle()
)

type DataMsg []models.Connection

type model struct {
	focusIndex  int
	inputs      []textinput.Model
	connections []models.Connection
	loading     bool
}

func InitialModel() model {
	m := model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CharLimit = 32
		t.Width = 64

		switch i {
		case 0:
			t.Placeholder = "From"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "To"
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			m.loading = true
			return m, m.searchCmd()

		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}

	case DataMsg:
		m.loading = false
		m.connections = msg
		return m, nil
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	var b strings.Builder

	// Render inputs
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		b.WriteRune('\n')
	}

	// Render results
	if m.loading {
		b.WriteString(highlightStyle.Render(" Loading..."))
	} else {
		for _, c := range m.connections {
			depTime := c.FromData.Departure.Format("15:04")
			arrTime := c.ToData.Arrival.Format("15:04")

			line := fmt.Sprintf(" %s → %s  [%s]  %v transfers\n",
				depTime, arrTime, c.Duration, c.Transfers)
			b.WriteString(line)
		}
	}

	return b.String()
}

func (m model) searchCmd() tea.Cmd {
	return func() tea.Msg {
		from := m.inputs[0].Value()
		to := m.inputs[1].Value()

		results, err := api.FetchConnections(from, to)
		if err != nil {
			return nil
		}
		return DataMsg(results)
	}
}
