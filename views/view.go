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
	borderSize     = 2
	hdrHeight      = 3
	hdrMinWidth    = 82
	hdrElmtPadd    = 2
	rsltMrgn       = 1
	smplConnHeight = 9
	smplConnMrgn   = 3

	stopsLineFixedWidth = (borderSize * 2) + (smplConnMrgn * 2) + (2+5)*2 + 6
	stopsLineMinWidth   = 10

	fullConnPaddH = 3
	fullConnPaddV = 1
)

const (
	// Icons
	filledDot = "●"
	hollowDot = "○"
	horzLine  = "─"
	vertLine  = "│"

	arrIcon  = "󰗔"
	dptIcon  = ""
	pltIcon  = "󱀓"
	srchIcon = ""
	swpIcon  = ""
	vhcIcon  = ""
	wlkIcon  = ""
)

var (
	// Colors
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
	// Styles
	noStyle = lipgloss.NewStyle()

	focusedStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbRed).
			Padding(0, 1)

	blurredStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbMidGray).
			Padding(0, 1)

	detailedResultStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(sbbRed).
				Padding(fullConnPaddV, fullConnPaddH)

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

type DataMsg struct {
	connections []models.Connection
	err         error
}

type model struct {
	width, height int
	tabIndex      int
	resultIndex   int
	headerOrder   []focusable
	inputs        []textinput.Model
	isArrivalTime bool
	connections   []models.Connection
	loading       bool
	errorMsg      string
	searched      bool
}

func InitialModel() model {
	// Define input prompts
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
			t.CharLimit = 10
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
	// Define keymaps
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		inputWidth := (m.width - hdrElmtPadd - hdrMinWidth) / 2
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
			if err := m.validateInputs(); err != "" {
				m.errorMsg = err
				return m, nil
			}
			m.loading = true
			m.connections = nil
			m.errorMsg = ""
			m.searched = true
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
				if err := m.validateInputs(); err != "" {
					m.errorMsg = err
					return m, nil
				}
				m.loading = true
				m.connections = nil
				m.errorMsg = ""
				m.searched = true
				return m, m.searchCmd()
			}

		case "tab", "shift+tab":
			if msg.String() == "shift+tab" {
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
		if msg.err != nil {
			m.errorMsg = "Failed to fetch connections. Check your internet connection."
			return m, nil
		}
		m.connections = msg.connections
		m.resultIndex = 0
		if len(m.connections) == 0 {
			m.errorMsg = "No connections found for the specified route."
		}
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m model) View() string {
	header := m.renderHeader()
	results := lipgloss.JoinHorizontal(lipgloss.Top,
		noStyle.
			Height(m.resultsHeight()).
			Render(m.renderResults()),
		noStyle.
			Height(m.resultsHeight()).
			Render(m.renderDetailedResult()),
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		noStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(sbbDarkRed).
			Width(m.contentWidth()).
			Height(m.resultsHeight()).
			Padding(0, rsltMrgn).
			Render(results),
	)
}

func (m model) contentWidth() int {
	return max(m.width-hdrElmtPadd, 0)
}

func (m model) resultsHeight() int {
	return max(m.height-hdrHeight-hdrElmtPadd, 0)
}

func (m model) maxVisibleConnections() int {
	return max(m.resultsHeight()/smplConnHeight, 1)
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Check key input in input fields
		switch m.headerOrder[m.tabIndex].id {
		case "date":
			t := &m.inputs[2]
			s := msg.String()
			val := t.Value()

			if msg.Type == tea.KeyBackspace {
				if len(val) == 5 || len(val) == 8 {
					t.SetValue(val[:len(val)-1])
					return nil
				}
			}

			if len(s) == 1 && s >= "0" && s <= "9" {
				switch len(val) {
				case 0:
					if s > "2" {
						return nil
					}
				case 1:
					if val == "2" && s > "9" {
						return nil
					}
				case 2, 3:
				case 4:
					t.SetValue(val + "-" + s)
					t.SetCursor(len(val) + 2)
					return nil
				case 5:
					if val[5] == '0' && s == "0" {
						return nil
					}
					if val[5] == '1' && s > "2" {
						return nil
					}
				case 6:
				case 7:
					t.SetValue(val + "-" + s)
					t.SetCursor(len(val) + 2)
					return nil
				case 8:
					if val[8] == '0' && s == "0" {
						return nil
					}
					if val[8] == '3' && s > "1" {
						return nil
					}
				case 9:
				default:
					return nil
				}
			} else if msg.Type == tea.KeyRunes {
				return nil
			}

		case "time":
			t := &m.inputs[3]
			s := msg.String()
			val := t.Value()

			if msg.Type == tea.KeyBackspace && len(val) == 3 {
				t.SetValue(val[:1])
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
					t.SetValue(val + ":" + s)
					t.SetCursor(5)
					return nil
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

func (m model) validateInputs() string {
	if m.inputs[0].Value() == "" {
		return "Please enter a departure station."
	}
	if m.inputs[1].Value() == "" {
		return "Please enter an arrival station."
	}
	return ""
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
		return DataMsg{connections: res, err: err}
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
		icon = swpIcon
	case "isArrivalTime":
		if m.isArrivalTime {
			icon = arrIcon
		} else {
			icon = dptIcon
		}
	case "search":
		icon = srchIcon
	}
	return style.Render(icon)
}

func (m model) resultBoxWidth() int {
	return max((m.width-smplConnMrgn)/2, rsltMrgn+stopsLineMinWidth+stopsLineFixedWidth)
}

func (m model) renderResults() string {
	if m.loading {
		return "\n  Searching connections..."
	}

	if m.errorMsg != "" {
		return "\n  " + noStyle.Foreground(sbbRed).Render(m.errorMsg)
	}

	if len(m.connections) == 0 {
		if m.searched {
			return "\n  No connections found."
		}
		return "\n  Enter stations above to see timetables"
	}

	var boxes []string
	boxWidth := m.resultBoxWidth()

	for i, c := range m.connections {
		boxes = append(boxes, m.renderSimpleConnection(c, i, boxWidth))
	}

	return lipgloss.JoinVertical(lipgloss.Left, boxes...)
}

func (m model) renderDetailedResult() string {
	if len(m.connections) == 0 {
		return ""
	}

	boxWidth := m.width - borderSize*4 - m.resultBoxWidth()
	return m.renderFullConnection(m.connections[m.resultIndex], boxWidth)
}

func (m model) renderFullConnection(c models.Connection, width int) string {
	var lines []string
	innerWidth := width - borderSize - (fullConnPaddH * 2)

	for i, section := range c.Sections {
		isFirst := i == 0
		isLast := i == len(c.Sections)-1

		if section.Walk != nil {
			lines = append(lines, m.renderWalkSection(section)...)
		} else if section.Journey != nil {
			lines = append(lines, m.renderJourneySection(section, innerWidth, isFirst, isLast)...)
		}

		if !isLast {
			lines = append(lines, "", "")
		}
	}

	content := strings.Join(lines, "\n")
	boxHeight := m.resultsHeight() - borderSize - (fullConnPaddV * 2)
	return detailedResultStyle.Width(width).Height(boxHeight).Render(content)
}

func (m model) renderJourneySection(section models.Section, width int, isFirst, isLast bool) []string {
	var lines []string

	const timeCol = 5
	const delayCol = 4
	const symbolCol = 5
	const platformCol = 10

	depTime := section.Departure.Departure.Local().Format("15:04")
	depDelay := section.Departure.Delay
	depStation := section.Departure.Station.Name
	depPlatform := section.Departure.Platform

	depDot := hollowDot
	if isFirst {
		depDot = filledDot
	}

	depLine := m.formatStationLine(depTime, depDelay, depDot, depStation, depPlatform, width, timeCol, delayCol, symbolCol, true)
	lines = append(lines, depLine)

	indent := strings.Repeat(" ", timeCol+delayCol)
	spacingLine := fmt.Sprintf("%s  %s", indent, vertLine)
	lines = append(lines, spacingLine)

	vehicleIcon := noStyle.Background(sbbBlue).Foreground(sbbWhite).Render(" " + vhcIcon + " ")
	vehicleCategory := noStyle.Background(sbbRed).Foreground(sbbWhite).Bold(true).
		Render(section.Journey.Category + " " + section.Journey.Number)
	company := noStyle.Background(sbbWhite).Foreground(sbbBlack).
		Render(section.Journey.Operator)
	vehicleLine := fmt.Sprintf("%s  %s  %s %s %s", indent, vertLine, vehicleIcon, vehicleCategory, company)
	lines = append(lines, vehicleLine)

	destLine := fmt.Sprintf("%s  %s   → %s", indent, vertLine, section.Journey.To)
	lines = append(lines, destLine)

	lines = append(lines, spacingLine)

	arrTime := section.Arrival.Arrival.Local().Format("15:04")
	arrDelay := section.Arrival.Delay
	arrStation := section.Arrival.Station.Name
	arrPlatform := section.Arrival.Platform

	arrSymbol := vertLine
	if isLast {
		arrSymbol = filledDot
	}

	arrLine := m.formatStationLine(arrTime, arrDelay, arrSymbol, arrStation, arrPlatform, width, timeCol, delayCol, symbolCol, false)
	lines = append(lines, arrLine)

	return lines
}

func (m model) renderWalkSection(section models.Section) []string {
	var lines []string

	walkDuration := ""
	if section.Walk != nil {
		dur := section.Walk.Duration
		if dur > 0 {
			walkDuration = fmt.Sprintf("%d min", dur/60)
		} else {
			depTime := section.Departure.Departure.Time
			arrTime := section.Arrival.Arrival.Time
			if !depTime.IsZero() && !arrTime.IsZero() {
				walkDuration = fmt.Sprintf("%d min", int(arrTime.Sub(depTime).Minutes()))
			}
		}
	}

	walkLine := fmt.Sprintf("            %s", walkDuration)
	lines = append(lines, walkLine)

	return lines
}

func (m model) formatStationLine(timeStr string, delay int, symbol, station, platform string, width, timeCol, delayCol, symbolCol int, bold bool) string {
	textStyle := noStyle
	if bold {
		textStyle = noStyle.Bold(true)
	}

	timePart := textStyle.Render(timeStr)

	delayPart := ""
	if delay > 0 {
		delayStr := fmt.Sprintf("+%d", delay)
		delayPart = noStyle.Foreground(sbbRed).Bold(true).Render(fmt.Sprintf("%*s", delayCol, delayStr))
	} else {
		delayPart = strings.Repeat(" ", delayCol)
	}

	symbolPart := fmt.Sprintf("  %s  ", symbol)

	platformPart := ""
	platformVisibleLen := 0
	if platform != "" {
		platformPart = textStyle.Render(fmt.Sprintf("%s %s", pltIcon, platform))
		platformVisibleLen = len(platform) + 3
	}

	fixedWidth := timeCol + delayCol + symbolCol + platformVisibleLen
	availableForStation := max(width-fixedWidth-1, 5)

	truncatedStation := truncateString(station, availableForStation)
	stationPart := textStyle.Render(truncatedStation)

	stationLen := len(truncatedStation)
	padding := max(availableForStation-stationLen, 1)

	if platformPart != "" {
		return fmt.Sprintf("%s%s%s%s%s%s",
			timePart, delayPart, symbolPart, stationPart, strings.Repeat(" ", padding), platformPart)
	}
	return fmt.Sprintf("%s%s%s%s", timePart, delayPart, symbolPart, stationPart)
}

func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if maxLen <= 3 {
		return s[:min(len(s), maxLen)]
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (m model) renderSimpleConnection(c models.Connection, index int, width int) string {
	firstVehicle := 0
	for x := range c.Sections {
		if c.Sections[x].Journey != nil {
			firstVehicle = x
			break
		}
	}

	vehicleIcon := noStyle.Background(sbbBlue).Foreground(sbbWhite).Render(" " + vhcIcon + " ")
	vehicleCategory := noStyle.Background(sbbRed).Foreground(sbbWhite).Bold(true).
		Render(c.Sections[firstVehicle].Journey.Category + " " + c.Sections[firstVehicle].Journey.Number)
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
		platformOrWalk = pltIcon + " " + noStyle.Render(c.FromData.Platform)
	} else if c.Sections[0].Walk != nil {
		platformOrWalk = wlkIcon + " " + noStyle.Render(
			fmt.Sprintf("%vm", c.Sections[0].Arrival.Arrival.Sub(c.Sections[0].Departure.Departure).Minutes()),
		)
	}

	duration := noStyle.Render(formatDuration(c.Duration))

	bottomLinePadding := max(width-(borderSize*2+smplConnMrgn*2+smplConnMrgn*2+3+5), 1)

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
		strings.Repeat(" ", bottomLinePadding),
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
		return filledDot + horzLine + horzLine + filledDot
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
		return filledDot + strings.Repeat(horzLine+horzLine+hollowDot, c.Transfers) + horzLine + horzLine + filledDot
	}

	var sb strings.Builder
	sb.WriteString(filledDot)

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

		sb.WriteString(strings.Repeat(horzLine, lineChars))
		if i < len(sectionDurations)-1 {
			sb.WriteString(hollowDot)
		} else {
			sb.WriteString(filledDot)
		}
	}

	return sb.String()
}
