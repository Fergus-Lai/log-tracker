package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

const FOCUSED_OFFSET = 2
const EXPLORER_TAB = 0
const COMMAND_TAB = 1

func (m model) handleListInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keyPress := msg.String()
	if keyPress == "ctrl+c" {
		return m, tea.Quit
	}
	if keyPress == "ctrl+q" {
		m.lists.command.SetValue("")
		m.lists.command.Blur()
		m.lists.focusedTab = 1
		m.state = titleView
		return m, nil
	}
	newIndex := m.lists.focusedTab
	if keyPress == "tab" {
		newIndex++
	}
	if keyPress == "shift+tab" {
		newIndex--
	}
	if m.lists.focusedTab != newIndex {
		// Reset
		m.lists.Cursor = 0
		if m.lists.focusedTab == COMMAND_TAB {
			m.lists.command.Blur()
		}

		m.lists.focusedTab = newIndex % (m.lists.selectedCount + FOCUSED_OFFSET)

		// Set New
		if m.lists.focusedTab == COMMAND_TAB {
			m.lists.command.Focus()
		}
		return m, nil
	}

	// Explorer
	if m.lists.focusedTab == EXPLORER_TAB {
		switch keyPress {
		case "up":
			m.lists.Cursor = positiveMod(m.lists.Cursor-1, len(m.lists.selected))
		case "down":
			m.lists.Cursor = positiveMod(m.lists.Cursor+1, len(m.lists.selected))
		case "enter", "space":
			m.lists.selected[m.lists.Cursor] = !m.lists.selected[m.lists.Cursor]
			if m.lists.selected[m.lists.Cursor] {
				m.lists.selectedCount += 1
			} else {
				m.lists.selectedCount -= 1
			}
		}
	}

	// Command
	if m.lists.focusedTab == COMMAND_TAB {
		switch keyPress {
		case "enter":
			switch strings.Split(m.lists.command.Value(), " ")[0] {
			case "!exit", "!back", "!home", "!quit":
				m.lists.command.SetValue("")
				m.lists.command.Blur()
				m.lists.focusedTab = EXPLORER_TAB
				m.state = titleView
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.lists.command, cmd = m.lists.command.Update(msg)
		return m, cmd
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
			if i == m.focusedTab-FOCUSED_OFFSET {
				style = focusedStyle
			} else {
				style = boldStyle
			}
		}

		if m.focusedTab == EXPLORER_TAB {
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
		Height(height - 2).
		MaxHeight(height - 1).
		PaddingLeft(2)

	rightStyle := lipgloss.NewStyle().
		Width(4 * width / 5).
		Height(height - 2).
		PaddingLeft(2)

	commandStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(width)

	// Render components into their styled containers
	explorer := explorerStyle.Render(m.renderExplorer(files))
	right := rightStyle.Render("Right Content")
	command := commandStyle.Render(m.command.View())

	content := lipgloss.JoinHorizontal(lipgloss.Top, explorer, right)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Top, content, command))
}
