package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
)

var boldStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
var titleStyle = boldStyle.AlignHorizontal(lipgloss.Center)

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
			file: File{
				name: "",
				folderPath: "",
				fileNameString: "",
				data: []RowData{},
				formatter: "",
			},
		},
	}
}


func (m model) Init() tea.Cmd {
	// TODO: Read the files
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
				switch choice := m.title.choices[m.title.selected]; choice {
				case "Quit":
					return m, tea.Quit
				case "View Files":
					if (len(m.lists.lists) > 0) {
						m.state = listView
					} else {
						// TODO: Show warning
					}
				case "Add File":
					m.state = listView;
				}
			} 
		}
	}
	switch state := m.state; state {
	case titleView:


	// case listView:
	// 	return m, nil;
	// case inputView:
	// 	return m, nil;
	}
    // Return the updated model to the Bubble Tea runtime for processing.
    // Note that we're not returning a command.
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
		s += optionLine.String() + "\n"
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