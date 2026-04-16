package main

import "fmt"

func listLine(s string, isActive bool) string {
	if isActive {
		return (focusedStyle.Render(fmt.Sprintf("[x] %s", s))) + "\n\n"
	}
	return (blurredStyle.Render(fmt.Sprintf("[ ] %s", s))) + "\n\n"
}

func positiveMod(x int, n int) int {
	return ((x % n) + n) % n
}
