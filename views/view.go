// Package views
package views

import (
	"fmt"
	"strings"
	"time"

	"sbb-tui/api"
	"sbb-tui/models"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	sbbWhite    = lipgloss.Color("#F6F6F6")
	sbbMidGray  = lipgloss.Color("#333333")
	sbbDarkGray = lipgloss.Color("#212121")
	sbbBlack    = lipgloss.Color("#141414")
	sbbRed      = lipgloss.Color("#D82E20")
	sbbBlue     = lipgloss.Color("#2E3279")

	noStyle = lipgloss.NewStyle()

	focusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbRed).
			Padding(0, 1)

	blurredStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbDarkGray).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			MarginTop(1).
			Bold(true).
			Foreground(sbbWhite).
			Background(sbbRed)
)

type DataMsg []models.Connection

const (
	KindInput int = iota
	KindButton
)

type focusable struct {
	kind  int
	id    string
	index int
}

type model struct {
	width, height int
	focusIndex    int
	headerOrder   []focusable
	inputs        []textinput.Model
	connections   []models.Connection
	loading       bool
}

func InitialModel() model {
	m := model{
		headerOrder: []focusable{
			{KindInput, "from", 0},
			{KindInput, "to", 1},
			{KindButton, "swap", -1},
			{KindInput, "date", 2},
			{KindInput, "time", 3},
		},
		inputs: make([]textinput.Model, 4),
	}

	now := time.Now()

	for i := range m.inputs {
		t := textinput.New()
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "From"
			t.Focus()
			t.PromptStyle = lipgloss.NewStyle().Foreground(sbbRed)
		case 1:
			t.Placeholder = "To"
		case 2:
			t.Placeholder = now.Format("2006-01-02")
			t.Width = 12
		case 3:
			t.Placeholder = now.Format("15:04")
			t.Width = 7
		}
		m.inputs[i] = t
	}
	return m
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		inputWidth := (m.width / 4) - 2
		m.inputs[0].Width = inputWidth
		m.inputs[1].Width = inputWidth

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "q":
			if m.focusIndex > 1 {
				return m, tea.Quit
			}

		case "enter":
			active := m.headerOrder[m.focusIndex]

			if active.kind == KindButton {
				if active.id == "swap" {
					v1 := m.inputs[0].Value()
					m.inputs[0].SetValue(m.inputs[1].Value())
					m.inputs[1].SetValue(v1)
				}
				return m, nil
			}

			m.loading = true
			return m, m.searchCmd()

		case "tab", "shift+tab", "left", "right":
			if msg.String() == "left" || msg.String() == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.headerOrder) {
				m.focusIndex = 0
			}
			if m.focusIndex < 0 {
				m.focusIndex = len(m.headerOrder) - 1
			}

			var cmds []tea.Cmd
			for _, item := range m.headerOrder {
				if item.kind == KindInput {
					if item.index == m.headerOrder[m.focusIndex].index {
						cmds = append(cmds, m.inputs[item.index].Focus())
					} else {
						m.inputs[item.index].Blur()
					}
				}
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
	headerItem := func(idx int) string {
		item := m.headerOrder[idx]
		style := blurredStyle
		if m.focusIndex == idx {
			style = focusedStyle
		}

		if item.kind == KindInput {
			return style.Render(m.inputs[item.index].View())
		}

		icon := " "
		switch item.id {
		case "swap":
			icon = ""
		}
		return style.Render(icon)
	}

	var headerItems []string
	for i := range m.headerOrder {
		headerItems = append(headerItems, headerItem(i))
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, headerItems...)

	var results strings.Builder
	if m.loading {
		results.WriteString("\n  Searching connections...")
	} else if len(m.connections) == 0 {
		results.WriteString("\n  Enter stations above to see timetables")
	} else {
		for _, c := range m.connections {
			dep := c.FromData.Departure.Local().Format("15:04")
			arr := c.ToData.Arrival.Local().Format("15:04")

			// Duration cleanup
			parts := strings.Split(c.Duration, ":") // e.g. 00d01:15:00
			dur := parts[1] + " min"
			if len(parts[0]) > 3 && parts[0][3:] != "00" {
				dur = parts[0][3:] + "h " + parts[1] + "m"
			}

			fmt.Fprintf(&results, "\n %s  %s  %s  %s  (%v x)\n",
				noStyle.Bold(true).Render(dep),
				lipgloss.NewStyle().Foreground(sbbRed).Render("→"),
				noStyle.Bold(true).Render(arr),
				lipgloss.NewStyle().Foreground(sbbMidGray).Render(dur),
				c.Transfers,
			)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		focusedStyle.Width(m.width-2).Height(m.height-5).Render(results.String()),
	)
}

func (m model) searchCmd() tea.Cmd {
	return func() tea.Msg {
		res, err := api.FetchConnections(
			m.inputs[0].Value(),
			m.inputs[1].Value(),
			m.inputs[2].Value(),
			m.inputs[3].Value(),
		)
		if err != nil {
			return nil
		}
		return DataMsg(res)
	}
}
