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

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

var boldStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
var titleStyle = boldStyle.AlignHorizontal(lipgloss.Center)
var errorStyle = lipgloss.NewStyle().Foreground((lipgloss.Color("#ED4337")))

var DATA_PATH =  filepath.Join(".", "data")

// These imports will be used later in the tutorial. If you save the file
// now, Go might complain they are unused, but that's fine.
// You may also need to run `go mod tidy` to download bubbletea and its
// dependencies.

func initialModel() model {
	return model{
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
			File: File{
				Name: "",
				FolderPath: "",
				FileNameString: "",
				Data: []RowData{},
				Formatter: "",
			},
		},
	}
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
			if (m.state == titleView) {
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
					m.state = listView;
				}
			} 
		}
	}
    return m, nil
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
		return tea.NewView("Input")
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