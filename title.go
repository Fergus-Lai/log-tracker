package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

func (m *titleModel) render(width int, height int) tea.View {
	var s strings.Builder
	s.WriteString(boldStyle.Render("Log Viewer by Fergus"))
	s.WriteString("\n\n")
	for i, choice := range m.choices {
		if m.selected == i {
			s.WriteString(focusedStyle.Render(fmt.Sprintf("[x] %s", choice)))
		} else {
			fmt.Fprintf(&s, "[ ] %s", choice)
		}
		s.WriteString("\n\n")

	}

	if m.errorMessage != "" {
		s.WriteString(errorStyle.Render(m.errorMessage))
	}
	centeredContent := lipgloss.Place(
		width,
		height,
		lipgloss.Top,
		lipgloss.Left,
		s.String(),
	)
	return tea.NewView(centeredContent)
}

func (m *model) handleTitleInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "shift+tab":
		if m.state == titleView && m.title.selected > 0 {
			m.title.selected--
		}

	case "down", "tab":
		if m.state == titleView && m.title.selected < len(m.title.choices)-1 {
			m.title.selected++
		}

	case "enter", "space":
		m.title.errorMessage = ""
		switch choice := m.title.choices[m.title.selected]; choice {
		case "Quit":
			return m, tea.Quit
		case "View Logs":
			if len(m.lists.lists) > 0 {
				m.state = listView
			} else {
				m.title.errorMessage = "No file tracked, please add file"
			}
		case "Add Log Profile":
			m.state = inputView
		case "Edit Log Profile":
			m.state = editView
		}
	}
	return m, nil
}
