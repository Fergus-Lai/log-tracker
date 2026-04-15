package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	textinput "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

var (
	boldStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	titleStyle   = boldStyle.AlignHorizontal(lipgloss.Center)
	errorStyle   = lipgloss.NewStyle().Foreground((lipgloss.Color("#ED4337")))
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

var DATA_PATH = filepath.Join(".", "data")

// These imports will be used later in the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.

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

	var t textinput.Model
	for i := range m.edit.input.inputs {
		t = textinput.New()
		t.CharLimit = 256
		t.SetWidth(256)

		s := t.Styles()
		s.Cursor.Color = lipgloss.Color("205")
		s.Focused.Prompt = focusedStyle
		s.Focused.Text = focusedStyle
		s.Blurred.Prompt = blurredStyle
		s.Focused.Text = focusedStyle
		t.SetStyles(s)

		switch i {
		case 0:
			t.Placeholder = "Name"
			t.CharLimit = 64
			t.Focus()
		case 1:
			t.Placeholder = "Folder Path"
		case 2:
			t.Placeholder = "File Matching String"
		case 3:
			t.Placeholder = "Log Format Matcher (Regex)"
		}

		m.edit.input.inputs[i] = t
	}

	for i := range m.input.inputs {
		t = textinput.New()
		t.CharLimit = 256
		t.SetWidth(256)

		s := t.Styles()
		s.Cursor.Color = lipgloss.Color("205")
		s.Focused.Prompt = focusedStyle
		s.Focused.Text = focusedStyle
		s.Blurred.Prompt = blurredStyle
		s.Focused.Text = focusedStyle
		t.SetStyles(s)

		switch i {
		case 0:
			t.Placeholder = "Name"
			t.CharLimit = 64
			t.Focus()
		case 1:
			t.Placeholder = "Folder Path"
		case 2:
			t.Placeholder = "File Matching String"
		case 3:
			t.Placeholder = "Log Format Matcher (Regex)"
		}

		m.input.inputs[i] = t
	}
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
	return LoadData
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
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseRight && m.state == inputView {
			m.input.pasteCurrent()
		}
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
		}

	}
	return m, nil
}

func (m model) View() tea.View {
	switch m.state {
	case titleView:
		return m.title.render(m.width, m.height)
	case listView:
		return tea.NewView("List")
	case inputView:
		return m.input.render(m.width, m.height, false)
	case editView:
		if m.edit.editMode {
			return m.edit.input.render(m.width, m.height, true)
		}
		return m.edit.render(m.width, m.height, m.files)
	default:
		return tea.NewView("Loading...")
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
