package main

import (
	"fmt"
	"os"

	"sbb-tui/views"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := views.InitialModel()

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}
