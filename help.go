package main

import (
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
)

func (m Model) helpView() string {
	var keybinds []struct {
		key  string
		desc string
	}
	if m.returnMode == welcome {
		keybinds = []struct {
			key  string
			desc string
		}{
			{"↑ / k / ↓ / j", "Navigate sessions"},
			{"Enter", "Open selected session"},
			{"n", "Create new session"},
			{"d", "Open daily session"},
			{"r", "Rename selected session"},
			{"x", "Delete selected session"},
			{"?", "Show this help"},
			{"q / Ctrl+c", "Quit"},
		}
	} else {
		keybinds = []struct {
			key  string
			desc string
		}{
			{"h / ← / Shift+Tab", "Previous column"},
			{"l / → / Tab", "Next column"},
			{"Enter", "Move task to next column"},
			{"n", "New task"},
			{"x", "Delete selected task"},
			{"u", "Undo last action"},
			{"r", "Redo last undone action"},
			{"Ctrl+s", "Save session"},
			{"/", "Filter mode"},
			{"esc", "Back to session menu"},
			{"?", "Show this help"},
			{"q / Ctrl+c", "Quit"},
		}
	}

	var lines []string
	lines = append(lines, CardTitleStyle.Render("Keybinds"))
	lines = append(lines, "")
	for _, kb := range keybinds {
		key := HelpKeybindStyle.Render(kb.key)
		desc := HelpDescStyle.Render(kb.desc)
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, key, "  ", desc))
	}
	lines = append(lines, "")
	lines = append(lines, HelpStyle.Render("Press ? or Esc to close"))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	card := CardStyle.Render(content)
	return CenterIn(m.width, m.height, card)
}

func (m *Model) helpKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc", "enter":
		m.mode = m.returnMode
		return m, nil
	case "ctrl+c", "q":
		m.shutdown = true
		return m, tea.Quit
	}
	return m, nil
}
