package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	textinput "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/atotto/clipboard"
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
	m.activeIndex = 0
	m.inProgress = false

	// Clear all text inputs
	for i := range m.inputs {
		m.inputs[i].SetValue("")
		m.inputs[i].Blur()
	}

	// Refocus the first input specifically
	m.inputs[0].Focus()
}

func (m *model) handleInputViewUpdate(msg tea.KeyPressMsg, isEdit bool) (tea.Model, tea.Cmd) {
	movedIndex := false
	keyPress := msg.String()
	var inputMod *inputModel
	if isEdit {
		inputMod = &m.edit.input
	} else {
		inputMod = &m.input
	}
	switch keyPress {
	case "ctrl+v":
		inputMod.pasteCurrent()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "enter", "down", "tab":
		if inputMod.focusIndex < len(inputMod.inputs) {
			inputMod.focusIndex++
			movedIndex = true
		} else if keyPress != "down" {
			if inputMod.inProgress {
				return m, nil
			}
			switch inputMod.choices[inputMod.activeIndex] {
			case "Save":
				if slices.ContainsFunc(inputMod.inputs, func(n textinput.Model) bool { return n.Value() == "" }) {
					inputMod.errorMessage = "Missing field"
					return m, nil
				}
				inputMod.inProgress = true
				return m, m.saveInput(inputMod.inputs[0].Value(), inputMod.inputs[1].Value(), inputMod.inputs[2].Value(), inputMod.inputs[3].Value(), isEdit)
			case "Discard":
				inputMod.resetInput()
				if isEdit {
					m.edit.editMode = false
				} else {
					m.state = titleView
				}
				return m, nil
			case "Delete":
				inputMod.inProgress = true
				return m, m.delete()
			}
		}
	case "shift+tab", "up":
		if inputMod.focusIndex > 0 {
			inputMod.focusIndex--
			movedIndex = true
		}
	case "left":
		if inputMod.focusIndex == len(inputMod.inputs) {
			n := len(inputMod.choices)
			inputMod.activeIndex = ((inputMod.activeIndex-1)%n + n) % n
		}
	case "right":
		if inputMod.focusIndex == len(inputMod.inputs) {
			n := len(inputMod.choices)
			inputMod.activeIndex = ((inputMod.activeIndex+1)%n + n) % n
		}
	}
	if movedIndex {
		cmds := make([]tea.Cmd, len(inputMod.inputs))
		for i := 0; i <= len(inputMod.inputs)-1; i++ {
			if i == inputMod.focusIndex {
				cmds[i] = inputMod.inputs[i].Focus()
				continue
			}
			inputMod.inputs[i].Blur()
		}
		return m, tea.Batch(cmds...)
	}
	cmd := inputMod.updateInputs(msg)
	return m, cmd
}

func (m *inputModel) render(width int, height int, isEdit bool) tea.View {
	var b strings.Builder
	var c *tea.Cursor
	if isEdit {
		b.WriteString(boldStyle.Render("Edit Log Profile \n"))
	} else {
		b.WriteString(boldStyle.Render("Create Log Profile \n"))
	}
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

	b.WriteString("\n\n")
	for i, button := range m.choices {
		style := blurredStyle
		if i == m.activeIndex && m.focusIndex == len(m.inputs) {
			style = focusedStyle
		}
		b.WriteString(style.Render(fmt.Sprintf("  [%s]", button)))
	}

	if m.errorMessage != "" {
		b.WriteString(errorStyle.Render(m.errorMessage))
	}

	if m.inProgress && m.choices[m.activeIndex] == "Save" {
		b.WriteString("Saving...\n")
	}

	if m.inProgress && m.choices[m.activeIndex] == "Delete" {
		b.WriteString("Deleting...\n")
	}

	centeredContent := lipgloss.Place(
		width,
		height,
		lipgloss.Top,
		lipgloss.Left,
		b.String(),
	)
	v := tea.NewView(centeredContent)
	v.Cursor = c
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

type saveErrMsg struct {
	err    error
	isEdit bool
}
type fileSavedMsg struct {
	file   File
	isEdit bool
}

func (m *model) saveInput(name string, folderPath string, fileNameMatcher string, formatter string, isEdit bool) tea.Cmd {
	return func() tea.Msg {
		safeName := convertSafeName(name)

		for i, f := range m.files {
			if convertSafeName(f.Name) == safeName && (!isEdit || m.edit.selectedIndex != i) {
				return saveErrMsg{
					err:    errors.New("Duplicate Error"),
					isEdit: isEdit,
				}
			}
		}
		newFile := File{
			Name:           name,
			FolderPath:     folderPath,
			FileNameString: fileNameMatcher,
			Formatter:      formatter,
		}

		data, err := json.MarshalIndent(newFile, "", "  ")
		if err != nil {
			return saveErrMsg{
				err,
				isEdit,
			}
		}

		if isEdit && m.files[m.edit.selectedIndex].Name != name {
			err = os.Remove(filepath.Join(DATA_PATH, convertSafeName(m.files[m.edit.selectedIndex].Name)+".json"))
			if err != nil {
				return saveErrMsg{
					err,
					isEdit,
				}
			}
		}

		filePath := filepath.Join(DATA_PATH, safeName+".json")

		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			return saveErrMsg{
				err,
				isEdit,
			}
		}

		return fileSavedMsg{
			file:   newFile,
			isEdit: isEdit,
		}
	}
}

func convertSafeName(s string) string {
	return strings.ReplaceAll(strings.ToLower(s), " ", "_")
}

func (m *inputModel) pasteCurrent() {
	text, err := clipboard.ReadAll()
	if err != nil {
		return
	}
	m.inputs[m.focusIndex].SetValue(m.inputs[m.focusIndex].Value() + text)
}
