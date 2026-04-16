package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	spinner "charm.land/bubbles/v2/spinner"
	textinput "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

var (
	boldStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	errorStyle   = lipgloss.NewStyle().Foreground((lipgloss.Color("#ED4337")))
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

var DATA_PATH = filepath.Join(".", "data")

func initialModel() model {
	m := model{
		state: 0,
		title: titleModel{
			choices:  []string{"Add Log Profile", "Edit Log Profile", "View Logs", "Quit"},
			selected: 0,
		},
		lists: listsModel{
			lists:         []listModel{},
			selectedIndex: 0,
		},
		input: inputModel{
			inputs:      make([]textinput.Model, 4),
			choices:     []string{"Save", "Discard"},
			activeIndex: 0,
			inProgress:  false,
		},
		edit: editModel{
			input: inputModel{
				inputs:      make([]textinput.Model, 4),
				choices:     []string{"Save", "Discard", "Delete"},
				activeIndex: 0,
				inProgress:  false,
			},
			editMode:      false,
			selectedIndex: 0,
		},
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	m.spinner = s

	m.edit.input.inputs = initialInputModel(len(m.edit.input.inputs))
	m.input.inputs = initialInputModel(len(m.input.inputs))
	return m
}

func LoadData() tea.Msg {
	var loadedFiles []File

	// Ensure directory exists
	if _, err := os.Stat(DATA_PATH); errors.Is(err, fs.ErrNotExist) {
		os.Mkdir(DATA_PATH, os.ModePerm)
	}

	files, err := os.ReadDir(DATA_PATH)
	if err != nil {
		return dataLoadedMsg{err: err}
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		data, err := os.ReadFile(path.Join(DATA_PATH, file.Name()))
		if err != nil {
			continue
		}

		var f File
		if err := json.Unmarshal(data, &f); err != nil {
			continue
		}

		loadedFiles = append(loadedFiles, f)
	}

	return dataLoadedMsg{files: loadedFiles}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		LoadData,
		m.spinner.Tick,
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		if msg.err != nil {
			// Error handle
			return m, tea.Quit
		}
		m.files = msg.files
		listModels := make([]listModel, len(m.files))
		for i := range listModels {
			listModels[i] = listModel{
				Cursor: 0,
				Filter: Filter{
					searchString: "",
					regexOn:      false,
					level:        "",
				},
			}
		}
		m.lists.lists = listModels
		return m, nil
	case fileSavedMsg:
		if msg.isEdit {
			m.files[m.edit.selectedIndex] = msg.file
			m.edit.input.resetInput()
			m.edit.editMode = false
		} else {
			m.files = append(m.files, msg.file)
			m.lists.lists = append(m.lists.lists, listModel{
				Cursor: 0,
				Filter: Filter{
					searchString: "",
					regexOn:      false,
					level:        "",
				},
			})
			m.input.resetInput()
			m.state = titleView
		}
		return m, nil
	case saveErrMsg:
		var inputMod *inputModel
		if msg.isEdit {
			inputMod = &m.edit.input
		} else {
			inputMod = &m.input
		}
		switch msg.err.Error() {
		case "Duplicate Error":
			inputMod.errorMessage = "Profile with same name alreadt exists, please try again"
		case "Unable to delete":
			inputMod.errorMessage = "Unable to delete file, please try again"
		default:
			inputMod.errorMessage = "Unable to save file, please try again"
		}
		inputMod.inProgress = false
		return m, nil
	case fileDeletedMsg:
		m.edit.input.inProgress = false
		m.edit.input.resetInput()
		m.files = append(m.files[:msg.index], m.files[msg.index+1:]...)
		m.edit.editMode = false
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseRight && m.state == inputView {
			m.input.pasteCurrent()
		}
		return m, nil
	case tea.KeyPressMsg:
		switch m.state {
		case titleView:
			return m.handleTitleInput(msg)
		case inputView:
			return m.handleInputViewUpdate(msg, false)
		case editView:
			if m.edit.editMode {
				return m.handleInputViewUpdate(msg, true)
			}
			return m.handleEditViewUpdate(msg)
		case listView:
			return m, nil
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() tea.View {
	switch m.state {
	case titleView:
		return m.title.render(m.width, m.height)
	case listView:
		return tea.NewView("List")
	case inputView:
		return m.input.render(m.width, m.height, m.spinner.View(), "", false)
	case editView:
		if m.edit.editMode {
			return m.edit.input.render(m.width, m.height, m.spinner.View(), m.files[m.edit.selectedIndex].Name, true)
		}
		return m.edit.render(m.width, m.height, m.files)
	default:
		return tea.NewView(fmt.Sprintf("%s Loading...", m.spinner.View()))
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
