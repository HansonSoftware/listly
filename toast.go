package main

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// toastMsg is sent when the toast should disappear.
type toastMsg struct{}

// Toast is a transient overlay message shown in the bottom-left.
type Toast struct {
	text string
}

func toastTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg { return toastMsg{} })
}

// Render returns the toast's overlay string.
func (t *Toast) Render() string {
	return ToastStyle.Render(t.text)
}

// overlay composites top over base at (x, y).
func overlay(base, top string, x, y int) string {
	return lipgloss.NewCompositor(
		lipgloss.NewLayer(base),
		lipgloss.NewLayer(top).X(x).Y(y).Z(1),
	).Render()
}
