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

const (
	// Focusable item kinds
	KindInput int = iota
	KindButton
)

const (
	// Layout dimensions
	headerHeight        = 3
	resultBoxHeight     = 9
	layoutPadding       = 2
	borderSize          = 2
	headerFixedWidth    = 82
	resultBoxMargin     = 3
	stopsLineFixedWidth = (borderSize * 2) + (resultBoxMargin * 2) + (2+5)*2 + 6 // borderSizes + resultBoxMargins + (space+time)*2 + delays
	stopsLineMinWidth   = 10
)

var (
	sbbWhite      = lipgloss.Color("#FFFFFF")
	sbbMidWhite   = lipgloss.Color("#F6F6F6")
	sbbDarkWhite  = lipgloss.Color("#DDDDDD")
	sbbGray       = lipgloss.Color("#888888")
	sbbMidGray    = lipgloss.Color("#484848")
	sbbDarkGray   = lipgloss.Color("#333333")
	sbbLightBlack = lipgloss.Color("#212121")
	sbbBlack      = lipgloss.Color("#141414")
	sbbRed        = lipgloss.Color("#D82E20")
	sbbMidRed     = lipgloss.Color("#B52C24")
	sbbDarkRed    = lipgloss.Color("#862010")
	sbbLightBlue  = lipgloss.Color("#315086")
	sbbBlue       = lipgloss.Color("#2E3279")
	sbbGreen      = lipgloss.Color("#3A7446")
)

var (
	noStyle = lipgloss.NewStyle()

	focusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbRed).
			Padding(0, 1)

	blurredStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbMidGray).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbRed).
			Bold(true).
			Foreground(sbbWhite).
			Background(sbbRed)
)

type focusable struct {
	kind  int
	id    string
	index int
}

type DataMsg []models.Connection

type model struct {
	width, height int
	tabIndex      int
	resultIndex   int
	headerOrder   []focusable
	inputs        []textinput.Model
	isArrivalTime bool
	connections   []models.Connection
	loading       bool
}

func InitialModel() model {
	m := model{
		headerOrder: []focusable{
			{KindInput, "from", 0},
			{KindInput, "to", 1},
			{KindButton, "swap", -1},
			{KindButton, "isArrivalTime", -1},
			{KindInput, "date", 2},
			{KindInput, "time", 3},
			{KindButton, "search", -1},
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
			t.Prompt = " "
			t.Focus()
		case 1:
			t.Placeholder = "To"
			t.Prompt = " "
		case 2:
			t.Placeholder = now.Format("2006-01-02")
			t.Prompt = " "
			t.Width = 12
		case 3:
			t.Placeholder = now.Format("15:04")
			t.Prompt = " "
			t.Width = 7
			t.CharLimit = 5
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
		inputWidth := (m.width - layoutPadding - headerFixedWidth) / 2
		m.inputs[0].Width = inputWidth
		m.inputs[1].Width = inputWidth

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "q":
			active := m.headerOrder[m.tabIndex]
			if active.kind == KindButton {
				return m, tea.Quit
			}

		case "enter":
			m.loading = true
			return m, m.searchCmd()

		case " ":
			active := m.headerOrder[m.tabIndex]
			switch active.id {
			case "swap":
				tmp := m.inputs[0].Value()
				m.inputs[0].SetValue(m.inputs[1].Value())
				m.inputs[1].SetValue(tmp)
			case "isArrivalTime":
				m.isArrivalTime = !m.isArrivalTime
			case "search":
				m.loading = true
				return m, m.searchCmd()
			}

		case "tab", "shift+tab", "left", "right":
			if msg.String() == "left" || msg.String() == "shift+tab" {
				m.tabIndex--
			} else {
				m.tabIndex++
			}

			if m.tabIndex >= len(m.headerOrder) {
				m.tabIndex = 0
			}
			if m.tabIndex < 0 {
				m.tabIndex = len(m.headerOrder) - 1
			}

			var cmds []tea.Cmd
			for _, item := range m.headerOrder {
				if item.kind == KindInput {
					if item.index == m.headerOrder[m.tabIndex].index {
						cmds = append(cmds, m.inputs[item.index].Focus())
					} else {
						m.inputs[item.index].Blur()
					}
				}
			}
			return m, tea.Batch(cmds...)

		case "up":
			if len(m.connections) > 0 && m.resultIndex > 0 {
				m.resultIndex--
			}
		case "down":
			if len(m.connections) > 0 && m.resultIndex < len(m.connections)-1 {
				m.resultIndex++
			}
		}

	case DataMsg:
		m.loading = false
		m.connections = msg
		m.resultIndex = 0
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m model) View() string {
	header := m.renderHeader()
	results := m.renderResults()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		noStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbDarkRed).
			Width(m.contentWidth()).
			Height(m.resultsHeight()).
			Render(results),
	)
}

func (m model) contentWidth() int {
	return max(m.width-layoutPadding, 0)
}

func (m model) resultsHeight() int {
	return max(m.height-headerHeight-layoutPadding, 0)
}

func (m model) maxVisibleConnections() int {
	return max(m.resultsHeight()/resultBoxHeight, 1)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check key input in input fields
		switch m.headerOrder[m.tabIndex].id {
		case "time":
			t := &m.inputs[3]
			s := msg.String()
			val := t.Value()

			if msg.Type == tea.KeyBackspace && len(val) == 3 {
				t.SetValue(val[:1]) // Delete the colon AND the digit before it
				return nil
			}

			// Only process numeric runes for the following logic
			if len(s) == 1 && s >= "0" && s <= "9" {
				switch len(val) {
				// Logic for each digit
				case 0:
					if s > "2" {
						return nil
					}
				case 1:
					if val == "2" && s > "3" {
						return nil
					}
				// Add `:` when typing third digit
				case 2:
					if s >= "0" && s <= "9" {
						t.SetValue(val + ":" + s)
						t.SetCursor(5)
						return nil
					}
				case 3:
					if s > "5" {
						return nil
					}
				case 4:
				default:
					return nil
				}
			} else if msg.Type == tea.KeyRunes {
				return nil
			}
		}
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m model) searchCmd() tea.Cmd {
	maxConnections := m.maxVisibleConnections()
	return func() tea.Msg {
		res, err := api.FetchConnections(
			m.inputs[0].Value(),
			m.inputs[1].Value(),
			m.inputs[2].Value(),
			m.inputs[3].Value(),
			m.isArrivalTime,
			maxConnections,
		)
		if err != nil {
			return nil
		}
		return DataMsg(res)
	}
}

func (m model) renderHeader() string {
	var headerItems []string
	for i := range m.headerOrder {
		headerItems = append(headerItems, m.renderHeaderItem(i))
	}

	headerItems = append(headerItems, titleStyle.Render(" SBB TIMETABLES <+> "))

	return lipgloss.JoinHorizontal(lipgloss.Top, headerItems...)
}

func (m model) renderHeaderItem(idx int) string {
	item := m.headerOrder[idx]
	style := blurredStyle
	if m.tabIndex == idx {
		style = focusedStyle
	}

	if item.kind == KindInput {
		return style.Render(m.inputs[item.index].View())
	}

	icon := " "
	switch item.id {
	case "swap":
		icon = ""
	case "isArrivalTime":
		if m.isArrivalTime {
			icon = "󰗔"
		} else {
			icon = ""
		}
	case "search":
		icon = ""
	}
	return style.Render(icon)
}

func (m model) resultBoxWidth() int {
	return max((m.width-resultBoxMargin)/2, stopsLineMinWidth+stopsLineFixedWidth)
}

func (m model) renderResults() string {
	if m.loading {
		return "\n  Searching connections..."
	}

	if len(m.connections) == 0 {
		return "\n  Enter stations above to see timetables"
	}

	var boxes []string
	boxWidth := m.resultBoxWidth()

	for i, c := range m.connections {
		boxes = append(boxes, m.renderSimpleConnection(c, i, boxWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Left, boxes...)
}

func (m model) renderSimpleConnection(c models.Connection, index int, width int) string {
	firstVehicle := 0
	for x := range c.Sections {
		if c.Sections[x].Journey != nil {
			firstVehicle = x
			break
		}
	}

	vehicleIcon := noStyle.Background(sbbBlue).Foreground(sbbWhite).Render("  ")
	vehicleCategory := noStyle.Background(sbbRed).Foreground(sbbWhite).Bold(true).
		Render(c.Sections[firstVehicle].Journey.Category + c.Sections[firstVehicle].Journey.Number)
	company := noStyle.Background(sbbWhite).Foreground(sbbBlack).
		Render(c.Sections[firstVehicle].Journey.Operator)
	endStop := noStyle.Render(c.Sections[firstVehicle].Journey.To)

	dep := c.FromData.Departure.Local().Format("15:04")
	arr := c.ToData.Arrival.Local().Format("15:04")
	departure := noStyle.Bold(true).Render(dep)
	arrival := noStyle.Bold(true).Render(arr)

	departureDelay := formatDelay(c.Sections[firstVehicle].Departure.Delay)
	arrivalDelay := formatDelay(c.Sections[firstVehicle].Arrival.Delay)

	stopsLineWidth := max(width-stopsLineFixedWidth, stopsLineMinWidth)
	stopsLine := noStyle.Bold(true).Render(renderStopsLine(c, stopsLineWidth))

	platformOrWalk := ""
	if len(c.FromData.Platform) > 0 {
		platformOrWalk = "󱀓 " + noStyle.Render(c.FromData.Platform)
	} else if c.Sections[0].Walk != nil {
		platformOrWalk = ""
	}

	duration := noStyle.Render(formatDuration(c.Duration))

	content := fmt.Sprintf("\n  %s %s %s  %s\n\n  %s%s  %s  %s%s\n\n  %s%s%v\n",
		vehicleIcon,
		vehicleCategory,
		company,
		endStop,
		departure,
		departureDelay,
		stopsLine,
		arrival,
		arrivalDelay,
		platformOrWalk,
		strings.Repeat(" ", width-(borderSize*2+resultBoxMargin*2+resultBoxMargin*2+3+5)),
		duration,
	)

	style := blurredStyle.Width(width)
	if index == m.resultIndex {
		style = focusedStyle.Width(width)
	}

	return style.Render(content)
}

// 00d01:15:00" -> "1h 15m" or "15 min".
func formatDuration(duration string) string {
	parts := strings.Split(duration, ":")
	if len(parts) < 2 {
		return duration
	}

	minutes := parts[1]
	if len(parts[0]) > 3 && parts[0][3:] != "00" {
		hours := parts[0][3:]
		return hours + "h " + minutes + "m"
	}
	return minutes + "min"
}

func formatDelay(delay int) string {
	if delay > 0 {
		return noStyle.Foreground(sbbRed).Bold(true).Render(fmt.Sprintf(" +%d", delay))
	}
	return ""
}

func renderStopsLine(c models.Connection, totalWidth int) string {
	if len(c.Sections) == 0 {
		return "●──●"
	}

	var sectionDurations []time.Duration
	var totalSectionDuration time.Duration
	for _, s := range c.Sections {
		// Skip walking sections
		if s.Journey == nil {
			continue
		}
		dep := s.Departure.Departure.Time
		arr := s.Arrival.Arrival.Time
		if !dep.IsZero() && !arr.IsZero() {
			dur := arr.Sub(dep)
			sectionDurations = append(sectionDurations, dur)
			totalSectionDuration += dur
		}
	}

	if totalSectionDuration == 0 || len(sectionDurations) == 0 {
		// Fallback to equal distribution
		return "●" + strings.Repeat("──○", c.Transfers) + "──●"
	}

	var sb strings.Builder
	sb.WriteString("●")

	usedChars := 0
	for i, secDur := range sectionDurations {
		var lineChars int
		if i == len(sectionDurations)-1 {
			// Last section gets remaining chars to avoid rounding errors
			lineChars = totalWidth - usedChars
		} else {
			proportion := float64(secDur) / float64(totalSectionDuration)
			lineChars = int(proportion*float64(totalWidth) + 0.5)
		}
		lineChars = max(lineChars, 1)
		usedChars += lineChars

		sb.WriteString(strings.Repeat("─", lineChars))
		if i < len(sectionDurations)-1 {
			sb.WriteString("○")
		} else {
			sb.WriteString("●")
		}
	}

	return sb.String()
}

func parseDurationString(duration string) time.Duration {
	// Format: "00d01:15:00" -> 1h15m
	parts := strings.Split(duration, ":")
	if len(parts) < 3 {
		return 0
	}

	var days, hours, minutes, seconds int
	if strings.Contains(parts[0], "d") {
		fmt.Sscanf(parts[0], "%dd%d", &days, &hours)
	} else {
		fmt.Sscanf(parts[0], "%d", &hours)
	}
	fmt.Sscanf(parts[1], "%d", &minutes)
	fmt.Sscanf(parts[2], "%d", &seconds)

	return time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second
}
