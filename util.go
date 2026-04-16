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
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetWidth(256)
	tiStyle := ti.Styles()
	tiStyle.Cursor.Color = lipgloss.Color("205")
	tiStyle.Focused.Prompt = focusedStyle
	tiStyle.Focused.Text = focusedStyle
	tiStyle.Blurred.Prompt = blurredStyle
	tiStyle.Blurred.Text = blurredStyle
	return ti
}
