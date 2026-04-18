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
