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

var DATA_PATH =  filepath.Join(".", "data")

// These imports will be used later in the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.

func initialModel() model {
	m := model{
		state: 0,
		title:  titleModel{
			choices: []string{"Add File", "View Files","Quit"},
			selected: 0,
		},
		lists: listsModel{
			lists: []listModel{},
			selectedIndex: 0,
		},
		input: inputModel{
			inputs: make([]textinput.Model, 4),
			isSave: true,
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
        if file.IsDir() { continue }

        data, err := os.ReadFile(path.Join(DATA_PATH, file.Name()))
        if err != nil { continue }

        var f File
        if err := json.Unmarshal(data, &f); err != nil { continue }

        loadedLists = append(loadedLists, listModel{
            File:   f,
            Cursor: 0,
			Filter: Filter{
				searchString: "",
				regexOn: false,
				level: "",
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
				if (m.state == titleView && m.title.selected > 0) {
					m.title.selected--
				}

			case "right":
				if (m.state == titleView && m.title.selected < len(m.title.choices) - 1) {
					m.title.selected++
				}

			case "enter", "space":
				m.title.errorMessage = ""
				switch choice := m.title.choices[m.title.selected]; choice {
				case "Quit":
					return m, tea.Quit
				case "View Files":
					if (len(m.lists.lists) > 0) {
						m.state = listView
					} else {
						m.title.errorMessage = "No file tracked, please add file"
					}
				case "Add File":
					m.state = inputView;
				}
			}
		case inputView:
			var movedIndex = false
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter", "down", "tab":
				if m.input.focusIndex < len(m.input.inputs) {
					m.input.focusIndex++
					movedIndex = true
				} else if (msg.String() != "down") {
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

	}
    return m, nil
}

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


func (m model) View() tea.View {
    switch m.state {
    case titleView:
		s := titleStyle.Render("Log Viewer by Fergus") + "\n\n";
		var optionLine strings.Builder;
		for i, choice := range m.title.choices {
			if (m.title.selected == i) {
				optionLine .WriteString(boldStyle.Render(fmt.Sprintf("[x] %s  ", choice)))
			} else {
				fmt.Fprintf(&optionLine, "[ ] %s  ", choice)
			}
			
			
		}
		s += optionLine.String() + "\n\n"
		if (m.title.errorMessage != "") {
			s += errorStyle.Render(m.title.errorMessage)
		}
		centeredContent := lipgloss.Place(
			m.width,   // The total width of your terminal
			m.height,  // The total height of your terminal
			lipgloss.Center,
			lipgloss.Center,
			s,
		)
        return tea.NewView(centeredContent)
    case listView:
        return tea.NewView("List")
	case inputView:
		var b strings.Builder
		var c *tea.Cursor
		b.WriteString(boldStyle.Render("Create Log Profile \n"))
		b.WriteRune('\n')
		for i, in := range m.input.inputs {
			b.WriteString(m.input.inputs[i].View())
			if i < len(m.input.inputs)-1 {
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
		if m.input.focusIndex == len(m.input.inputs) && m.input.isSave {
			saveButton = &focusedSave
		}
		discardButton := &blurredDiscard
		if m.input.focusIndex == len(m.input.inputs) && !m.input.isSave {
			discardButton = &focusedDiscard
		}
		fmt.Fprintf(&b, "\n\n%s  %s\n\n", *saveButton, *discardButton)


		centeredContent := lipgloss.Place(
			m.width,   // The total width of your terminal
			m.height,  // The total height of your terminal
			lipgloss.Top,
			lipgloss.Left,
			b.String(),
		)
		v := tea.NewView(centeredContent)
		v.Cursor = c
		return v
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