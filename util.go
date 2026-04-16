package main

import (
	"fmt"

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

func initTextInput(placeholder string) textinput.Model {
	t := textinput.New()
	t.Placeholder = placeholder
	t.SetWidth(256)
	s := t.Styles()
	s.Cursor.Color = lipgloss.Color("205")
	s.Focused.Prompt = focusedStyle
	s.Focused.Text = focusedStyle
	s.Blurred.Prompt = blurredStyle
	s.Blurred.Text = blurredStyle
	t.SetStyles(s)
	return t
}
