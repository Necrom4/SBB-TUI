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
	sbbWhite     = lipgloss.NewStyle()
	sbbDarkWhite = lipgloss.NewStyle().Foreground(lipgloss.Color("#F6F6F6"))
	sbbMidGray   = lipgloss.NewStyle().Foreground(lipgloss.Color("#333333"))
	sbbDarkGray  = lipgloss.NewStyle().Foreground(lipgloss.Color("#212121"))
	sbbBlack     = lipgloss.NewStyle().Foreground(lipgloss.Color("#141414"))
	sbbRed       = lipgloss.NewStyle().Foreground(lipgloss.Color("#D82E20"))
	sbbBlue      = lipgloss.NewStyle().Foreground(lipgloss.Color("#2E3279"))

	sbbTitle = lipgloss.NewStyle().MarginTop(1).Bold(true).Foreground(lipgloss.Color("#F6F6F6")).Background(lipgloss.Color("#D82E20"))
)

type DataMsg []models.Connection

type model struct {
	width, height int
	focusIndex    int
	inputs        []textinput.Model
	swapBtn       string
	connections   []models.Connection
	loading       bool
}

func InitialModel() model {
	m := model{
		inputs:  make([]textinput.Model, 2),
		swapBtn: "<->",
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.CharLimit = 36
		t.Width = 64

		switch i {
		case 0:
			t.Placeholder = "From"
			t.Focus()
			t.PromptStyle = sbbRed
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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.inputs[0].Width = (m.width / 2) - 6
		m.inputs[1].Width = (m.width / 2) - 6
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "q":
			if m.focusIndex > 1 {
				return m, tea.Quit
			}

		case "enter":
			if m.focusIndex == 2 {
				fromVal := m.inputs[0].Value()
				toVal := m.inputs[1].Value()
				m.inputs[0].SetValue(toVal)
				m.inputs[1].SetValue(fromVal)
				return m, nil
			}
			m.loading = true
			return m, m.searchCmd()

		case "tab", "shift+tab", "left", "right":
			s := msg.String()

			if s == "left" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex >= len(m.inputs)+1 {
				m.focusIndex--
			} else if m.focusIndex < 0 {
				m.focusIndex++
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = sbbRed
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = sbbWhite
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
	if m.width == 0 {
		return "Initializing..."
	}

	headerHeight := 3
	resultsHeight := m.height - headerHeight - 3 // Remaining space

	columnWidth := m.width / 4

	inputBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(sbbMidGray.GetForeground()).
		Width(columnWidth - 1).
		Height(1)

	swapBtnStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(sbbMidGray.GetForeground())

	resultsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(sbbRed.GetForeground()).
		Width(m.width-2).
		Height(resultsHeight).
		Padding(0, 1)

	fromStyle := inputBoxStyle
	if m.focusIndex == 0 {
		fromStyle = fromStyle.BorderForeground(sbbRed.GetForeground())
	}
	fromView := fromStyle.Render(m.inputs[0].View())

	toStyle := inputBoxStyle
	if m.focusIndex == 1 {
		toStyle = toStyle.BorderForeground(sbbRed.GetForeground())
	}
	toView := toStyle.Render(m.inputs[1].View())

	if m.focusIndex == 2 {
		swapBtnStyle = swapBtnStyle.BorderForeground(sbbRed.GetForeground())
	}
	swapBtnView := swapBtnStyle.Render("")

	// ↮
	header := lipgloss.JoinHorizontal(lipgloss.Top, fromView, toView, swapBtnView, " ", sbbTitle.Render(" SBB TIMETABLES <+> "))

	var results strings.Builder
	if m.loading {
		results.WriteString("\n  Searching connections...")
	} else if len(m.connections) == 0 {
		results.WriteString("\n  Enter stations above to see timetables")
	} else {
		for _, c := range m.connections {
			dep := c.FromData.Departure.Local().Format("15:04")
			arr := c.ToData.Arrival.Local().Format("15:04")

			durRaw := strings.Split(c.Duration, ":")
			dur := durRaw[1] + " min"
			if durRaw[0] != "00d00" {
				dur = durRaw[0][3:] + "h " + durRaw[1] + "m"
			}

			fmt.Fprintf(&results, "\n %s  %s  %s  %s  (%v x)\n",
				sbbWhite.Bold(true).Render(dep),
				sbbRed.Render("→"),
				sbbWhite.Bold(true).Render(arr),
				sbbMidGray.Render(dur),
				c.Transfers,
			)
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		resultsStyle.Render(results.String()),
	)
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
