package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)


func (m *inputModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m *inputModel) resetInput() {
    // Reset the focus to the first element
    m.focusIndex = 0
    m.isSave = true

    // Clear all text inputs
    for i := range m.inputs {
        m.inputs[i].SetValue("")
        m.inputs[i].Blur()
    }
    
    // Refocus the first input specifically
    m.inputs[0].Focus()
}

func (m *model) handleInputViewUpdate(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	movedIndex := false
	keyPress := msg.String()
	switch keyPress {
	case "ctrl+c":
		return m, tea.Quit
	case "enter", "down", "tab":
		if m.input.focusIndex < len(m.input.inputs) {
			m.input.focusIndex++
			movedIndex = true
		} else if (keyPress != "down") {
			if (m.input.isSave) {

			} else {
				m.input.resetInput() // <--- Reset the fields
				m.state = titleView
				return m, nil
			}
		}
	case "shift+tab", "up":
		if m.input.focusIndex > 0 {
			m.input.focusIndex--
			movedIndex = true
		}
	case "left", "right":
		if m.input.focusIndex == len(m.input.inputs) {
			m.input.isSave = !m.input.isSave
		}
	}
	if (movedIndex) {
		cmds := make([]tea.Cmd, len(m.input.inputs))
		for i := 0; i <= len(m.input.inputs)-1; i++ {
			if i == m.input.focusIndex {
				cmds[i] = m.input.inputs[i].Focus()
				continue
			}
			m.input.inputs[i].Blur()
		}
		return m, tea.Batch(cmds...)
	}
	cmd := m.input.updateInputs(msg)
	return m, cmd
}


func (m *inputModel) render(width int, height int) tea.View {
	var b strings.Builder
	var c *tea.Cursor
	b.WriteString(boldStyle.Render("Create Log Profile \n"))
	b.WriteRune('\n')
	for i, in := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
			b.WriteRune('\n')
		}
		if in.Focused() {
			c = in.Cursor()
			if c != nil {
				c.Y += i
			}
		}
	}

	saveButton := &blurredSave
	if m.focusIndex == len(m.inputs) && m.isSave {
		saveButton = &focusedSave
	}
	discardButton := &blurredDiscard
	if m.focusIndex == len(m.inputs) && !m.isSave {
		discardButton = &focusedDiscard
	}
	fmt.Fprintf(&b, "\n\n%s  %s\n\n", *saveButton, *discardButton)


	centeredContent := lipgloss.Place(
		width,   
		height, 
		lipgloss.Top,
		lipgloss.Left,
		b.String(),
	)
	v := tea.NewView(centeredContent)
	v.Cursor = c
	return v
}