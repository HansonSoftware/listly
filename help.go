package main

import (
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
)

func (m Model) helpView() string {
	keybinds := welcomeKeybinds
	if m.returnMode != welcome {
		keybinds = boardKeybinds
	}

	var lines []string
	lines = append(lines, CardTitleStyle.Render("Keybinds"))
	lines = append(lines, "")
	for _, b := range keybinds {
		key := HelpKeybindStyle.Render(b.Help().Key)
		desc := HelpDescStyle.Render(b.Help().Desc)
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
		m.setMode(m.returnMode)
		return m, nil
	case "ctrl+c", "q":
		m.shutdown = true
		return m, tea.Quit
	}
	return m, nil
}
