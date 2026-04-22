package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

const FOCUSED_OFFSET = 3
const EXPLORER_TAB = 0
const FILTER_TAB = 1
const COMMAND_TAB = 2

func (m model) handleListInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keyPress := msg.String()
	if keyPress == "ctrl+c" {
		return m, tea.Quit
	}
	if keyPress == "ctrl+q" {
		m.lists.command.SetValue("")
		m.lists.command.Blur()
		if m.lists.focusedTab == FILTER_TAB && m.lists.Cursor != 2 {
			m.lists.Filter.inputs[m.lists.Cursor].Blur()
		}
		m.lists.focusedTab = EXPLORER_TAB
		m.state = titleView
		return m, nil
	}
	newIndex := m.lists.focusedTab
	if keyPress == "ctrl+r" {
		newIndex = COMMAND_TAB
	}
	if keyPress == "tab" {
		newIndex++
	}
	if keyPress == "shift+tab" {
		newIndex--
	}
	if m.lists.focusedTab != newIndex {
		// Reset
		if m.lists.focusedTab == COMMAND_TAB {
			m.lists.command.Blur()
		}
		if m.lists.focusedTab == FILTER_TAB && m.lists.Cursor != 2 {
			m.lists.Filter.inputs[m.lists.Cursor].Blur()
		}
		m.lists.Cursor = 0

		m.lists.focusedTab = newIndex % (m.lists.selectedCount + FOCUSED_OFFSET)

		// Set New
		if m.lists.focusedTab == COMMAND_TAB {
			m.lists.command.Focus()
		}
		if m.lists.focusedTab == FILTER_TAB {
			m.lists.Filter.inputs[0].Focus()
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

	if m.lists.focusedTab == FILTER_TAB {
		newCursor := m.lists.Cursor
		switch keyPress {
		case "enter", "down":
			newCursor = positiveMod(m.lists.Cursor+1, len(m.lists.Filter.inputs)+1)
		case "up":
			newCursor = positiveMod(m.lists.Cursor-1, len(m.lists.Filter.inputs)+1)
		case "space":
			if m.lists.Cursor == 2 {
				m.lists.Filter.regexOn = !m.lists.Filter.regexOn
				return m, nil
			}
		}
		if newCursor != m.lists.Cursor {
			if m.lists.Cursor < len(m.lists.Filter.inputs) {
				m.lists.Filter.inputs[m.lists.Cursor].Blur()
			}
			if newCursor < len(m.lists.Filter.inputs) {
				m.lists.Filter.inputs[newCursor].Focus()
			}
			m.lists.Cursor = newCursor
			return m, nil
		}
		var cmd tea.Cmd
		m.lists.Filter.inputs[m.lists.Cursor], cmd = m.lists.Filter.inputs[m.lists.Cursor].Update(msg)
		return m, cmd

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
				return m, nil
			default:
				m.lists.command.SetValue("")
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.lists.command, cmd = m.lists.command.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *listsModel) renderExplorer(files FilesMap) string {
	var b strings.Builder
	titleStyle := boldStyle
	if m.focusedTab == EXPLORER_TAB {
		titleStyle = focusedStyle
	}
	b.WriteString(titleStyle.Render("Explorer"))
	b.WriteString("\n\n")
	j := 0
	for i, f := range files.slice {
		cursor := "[ ]"
		style := blurredStyle
		if m.selected[i] {
			cursor = "[x]"
			if j == m.focusedTab-FOCUSED_OFFSET {
				style = boldStyle
			} else {
				style = blurredStyle.Bold(true)
			}
			j++
		}

		if m.focusedTab == EXPLORER_TAB {
			if m.Cursor == i {
				style = focusedStyle
			}
		}

		b.WriteString(style.Render(fmt.Sprintf("  %s %s", cursor, f)))
		b.WriteString("\n\n")
	}
	return b.String()
}

func (m *listsModel) renderFilter(width int) string {
	cellLeftStyle := boldStyle

	titleStyle := boldStyle
	if m.focusedTab == FILTER_TAB {
		titleStyle = focusedStyle
	}
	titleCellStyle := titleStyle.Width(width)

	var b strings.Builder

	b.WriteString(titleCellStyle.Render("Filter"))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cellLeftStyle.Render("Level: "), m.Filter.inputs[0].View()))
	b.WriteString("\n\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cellLeftStyle.Render("Word: "), m.Filter.inputs[1].View()))
	b.WriteString("\n\n")
	regexOn := "False"
	style := boldStyle
	if m.focusedTab == FILTER_TAB && m.Cursor == len(m.Filter.inputs) {
		style = focusedStyle
	}
	if m.Filter.regexOn {
		regexOn = "True"
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, cellLeftStyle.Render("Regex On: "), style.Render(regexOn)))

	return b.String()
}

func (m *listsModel) render(width int, height int, files FilesMap) tea.View {
	explorerStyle := lipgloss.NewStyle().
		Width(width / 4).
		Height(2 * (height - 2) / 3).
		MaxHeight(2 * (height - 2) / 3).
		PaddingLeft(2)

	filterStyle := lipgloss.NewStyle().
		Width(width/4).
		Height((height-2)/3).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		MaxHeight((height - 2) / 3).
		PaddingLeft(2)

	rightStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Width(3 * width / 4).
		Height(height - 2).
		PaddingLeft(2)

	commandStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(width)

	// Render components into their styled containers
	explorer := explorerStyle.Render(m.renderExplorer(files))
	filter := filterStyle.Render(m.renderFilter(width / 5))

	left := lipgloss.JoinVertical(lipgloss.Top, explorer, filter)
	right := rightStyle.Render("Right Content")
	command := commandStyle.Render(m.command.View())

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Top, content, command))
}
