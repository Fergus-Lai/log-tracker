package main

import (
	"errors"
	"fmt"
	"slices"

	textinput "charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"
)

func listLine(s string, isActive bool) string {
	if isActive {
		return (focusedStyle.Render(fmt.Sprintf("[x] %s", s))) + "\n\n"
	}
	return (blurredStyle.Render(fmt.Sprintf("[ ] %s", s))) + "\n\n"
}

func positiveMod(x int, n int) int {
	return ((x % n) + n) % n
}

func initTextInput(placeholder string, focusStyle lipgloss.Style, blurStyle lipgloss.Style) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.SetWidth(256)
	s := t.Styles()
	s.Cursor.Color = lipgloss.Color("205")
	s.Focused.Prompt = focusStyle
	s.Focused.Text = focusStyle
	s.Blurred.Prompt = blurStyle
	s.Blurred.Text = blurStyle
	t.SetStyles(s)
	return t
}

type FilesMap struct {
	slice []string
	data  map[string]File
}

func NewFilesMap() *FilesMap {
	return &FilesMap{
		slice: []string{},
		data:  make(map[string]File),
	}
}

func (m *FilesMap) AddFile(f File) error {
	_, hasData := m.data[f.Name]
	if hasData {
		return errors.New("Duplicate name")
	}
	m.slice = append(m.slice, f.Name)
	m.data[f.Name] = f
	return nil
}

func (m *FilesMap) RemoveFile(i int) error {
	_, hasData := m.data[m.slice[i]]
	if !hasData {
		return errors.New("File not found")
	}
	delete(m.data, m.slice[i])
	m.slice = slices.Delete(m.slice, i, i+1)
	return nil
}
