package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	textinput "charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

var boldStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
var titleStyle = boldStyle.AlignHorizontal(lipgloss.Center)
var errorStyle = lipgloss.NewStyle().Foreground((lipgloss.Color("#ED4337")))

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedSave = focusedStyle.Render("[ Save ]")
	blurredSave = fmt.Sprintf("[ %s ]", blurredStyle.Render("Save"))

	focusedDiscard = focusedStyle.Render("[ Discard ]")
	blurredDiscard = fmt.Sprintf("[ %s ]", blurredStyle.Render("Discard"))
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
			choices:  []string{"Add File", "View Files", "Quit"},
			selected: 0,
		},
		lists: listsModel{
			lists:         []listModel{},
			selectedIndex: 0,
		},
		input: inputModel{
			inputs: make([]textinput.Model, 4),
			isSave: true,
			saving: false,
		},
	}

	var t textinput.Model
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
	var loadedLists []listModel

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

		loadedLists = append(loadedLists, listModel{
			File:   f,
			Cursor: 0,
			Filter: Filter{
				searchString: "",
				regexOn:      false,
				level:        "",
			},
		})
	}

	return dataLoadedMsg{lists: loadedLists}
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
		m.lists.lists = msg.lists
	case fileSavedMsg:
		m.lists.lists = append(m.lists.lists, listModel{
			File: msg.file,
		})
		m.input.resetInput()
		m.state = titleView
		return m, nil
	case saveErrMsg:
		if msg.err.Error() == "Duplicate Error" {
			m.input.errorMessage = "Profile with same name alreadt exists, please try again"
		} else {
			m.input.errorMessage = "Unable to save file, please try again"
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch m.state {
		case titleView:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit

			case "left":
				if m.state == titleView && m.title.selected > 0 {
					m.title.selected--
				}

			case "right":
				if m.state == titleView && m.title.selected < len(m.title.choices)-1 {
					m.title.selected++
				}

			case "enter", "space":
				m.title.errorMessage = ""
				switch choice := m.title.choices[m.title.selected]; choice {
				case "Quit":
					return m, tea.Quit
				case "View Files":
					if len(m.lists.lists) > 0 {
						m.state = listView
					} else {
						m.title.errorMessage = "No file tracked, please add file"
					}
				case "Add File":
					m.state = inputView
				}
			}
		case inputView:
			return m.handleInputViewUpdate(msg)
		}

	}
	return m, nil
}

func (m model) View() tea.View {
	switch m.state {
	case titleView:
		s := titleStyle.Render("Log Viewer by Fergus") + "\n\n"
		var optionLine strings.Builder
		for i, choice := range m.title.choices {
			if m.title.selected == i {
				optionLine.WriteString(boldStyle.Render(fmt.Sprintf("[x] %s  ", choice)))
			} else {
				fmt.Fprintf(&optionLine, "[ ] %s  ", choice)
			}

		}
		s += optionLine.String() + "\n\n"
		if m.title.errorMessage != "" {
			s += errorStyle.Render(m.title.errorMessage)
		}
		centeredContent := lipgloss.Place(
			m.width,  // The total width of your terminal
			m.height, // The total height of your terminal
			lipgloss.Center,
			lipgloss.Center,
			s,
		)
		return tea.NewView(centeredContent)
	case listView:
		return tea.NewView("List")
	case inputView:
		return m.input.render(m.width, m.height)
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
