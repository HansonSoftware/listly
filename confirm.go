package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ConfirmDialog is a shared "are you sure?" card used by every
// destructive action (task delete, session delete).
type ConfirmDialog struct {
	prompt    string
	onConfirm func() tea.Cmd
	onCancel  func()
	width     int
	height    int
}

func NewConfirmDialog(prompt string, onConfirm func() tea.Cmd, onCancel func()) *ConfirmDialog {
	return &ConfirmDialog{prompt: prompt, onConfirm: onConfirm, onCancel: onCancel}
}

func (c *ConfirmDialog) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "x":
			if c.onConfirm != nil {
				return c.onConfirm()
			}
		default:
			if c.onCancel != nil {
				c.onCancel()
			}
		}
	}
	return nil
}

func (c *ConfirmDialog) View() tea.View {
	content := lipgloss.JoinVertical(lipgloss.Left,
		CardTitleStyle.Render("Are you sure?"),
		"",
		c.prompt,
		"",
		HelpStyle.Render("enter/x: confirm  •  any other key: cancel"),
	)
	v := tea.NewView(CenterIn(c.width, c.height, CardStyle.Render(content)))
	v.AltScreen = true
	v.WindowTitle = "Listly"
	return v
}
