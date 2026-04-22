package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

func (m *editModel) render(width int, height int, files FilesMap) tea.View {
	var b strings.Builder
	b.WriteString(boldStyle.Render("Edit Log Profile \n"))
	b.WriteRune('\n')

	for i, f := range files.slice {
		b.WriteString(listLine(f, m.selectedIndex == i))
	}

	b.WriteString(listLine("Return to home screen", m.selectedIndex == len(files.slice)))

	centeredContent := lipgloss.Place(
		width,
		height,
		lipgloss.Top,
		lipgloss.Left,
		b.String(),
	)
	return tea.NewView(centeredContent)
}

func (m *model) handleEditViewUpdate(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	keyPress := msg.String()
	switch keyPress {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "shift+tab":
		m.edit.selectedIndex = positiveMod(m.edit.selectedIndex-1, len(m.files.slice)+1)
	case "down", "tab":
		m.edit.selectedIndex = positiveMod(m.edit.selectedIndex+1, len(m.files.slice)+1)

	case "enter":
		if m.edit.selectedIndex == len(m.files.slice) {
			m.state = titleView
			m.edit.selectedIndex = 0
		} else {
			// keys := make([]string, 0, len(m.files.slice))
			// for k := range m.files {
			// 	keys = append(keys, k)
			// }

			f := m.files.data[m.files.slice[m.edit.selectedIndex]]
			for i := range m.edit.input.inputs {
				switch i {
				case 0:
					m.edit.input.inputs[i].SetValue(f.Name)
					m.edit.input.inputs[i].Focus()
				case 1:
					m.edit.input.inputs[i].SetValue(f.FolderPath)
				case 2:
					m.edit.input.inputs[i].SetValue(f.FileNameString)
				case 3:
					m.edit.input.inputs[i].SetValue(f.Formatter)
				}
			}
			m.edit.editMode = true
		}

	}
	return m, nil
}

type fileDeletedMsg struct {
	index int
}

func (m *model) delete() tea.Cmd {
	return func() tea.Msg {
		safeName := convertSafeName(m.files.slice[m.edit.selectedIndex])
		filePath := filepath.Join(DATA_PATH, safeName+".json")
		err := os.Remove(filePath)
		if err != nil {
			return saveErrMsg{
				err:    errors.New("Unable to delete"),
				isEdit: true,
			}
		}
		return fileDeletedMsg{
			index: m.edit.selectedIndex,
		}
	}
}
