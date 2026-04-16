package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

func (m model) handleListInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keyPress := msg.String()
	if keyPress == "ctrl+c" {
		return m, tea.Quit
	}
	if m.lists.focusedTab == -1 {
		switch keyPress {
		case "up":
			m.lists.Cursor = positiveMod(m.lists.Cursor-1, len(m.lists.selected))
		case "down":
			m.lists.Cursor = positiveMod(m.lists.Cursor+1, len(m.lists.selected))
		case "enter":
			m.lists.selected[m.lists.Cursor] = !m.lists.selected[m.lists.Cursor]
		}
	}

	return m, nil
}

func (m *listsModel) renderExplorer(files []File) string {
	var b strings.Builder
	b.WriteString(boldStyle.Render("Explorer"))
	b.WriteString("\n\n")
	for i, f := range files {
		cursor := "[ ]"
		style := blurredStyle
		if m.selected[i] {
			cursor = "[x]"
			if i == m.focusedTab {
				style = focusedStyle
			} else {
				style = boldStyle
			}
		}

		if m.focusedTab == -1 {
			if m.Cursor == i {
				style = focusedStyle
			}
		}

		b.WriteString(style.Render(fmt.Sprintf("  %s %s", cursor, f.Name)))
		b.WriteString("\n\n")
	}
	return b.String()
}

func (m *listsModel) render(width int, height int, files []File) tea.View {
	explorerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, true, false, false).
		Width(width / 5).
		Height(height).
		MaxHeight(height).
		MarginLeft(2)

	rightStyle := lipgloss.NewStyle().
		Width(4 * width / 5).
		Height(height).
		MarginLeft(2)

	// Render components into their styled containers
	explorer := explorerStyle.Render(m.renderExplorer(files))
	right := rightStyle.Render("Right Content")

	return tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, explorer, right))
}
