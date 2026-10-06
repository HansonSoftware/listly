package main

import (
	"fmt"

	bubbleshelp "charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// welcomeKeybinds is the single source of truth for session-menu keybinds:
// the footer help component and the ? view both render from this slice.
var welcomeKeybinds = []key.Binding{
	key.NewBinding(key.WithKeys("up", "k", "down", "j"), key.WithHelp("↑/↓/j/k", "navigate")),
	key.NewBinding(key.WithKeys("g"), key.WithHelp("g/G", "top/bottom")),
	key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "daily")),
	key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "rename")),
	key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
	key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more")),
	key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

// welcomeFooter implements help.KeyMap for the session menu footer.
type welcomeFooter struct{}

func (welcomeFooter) ShortHelp() []key.Binding {
	return filterBindings(welcomeKeybinds, "navigate", "select", "more")
}
func (welcomeFooter) FullHelp() [][]key.Binding { return [][]key.Binding{welcomeKeybinds} }

func (m Model) welcomeView() string {
	if len(m.sessions) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Center,
			LogoStyle.Render(Logo),
			WelcomeSubtitleStyle.Render("No sessions yet. Create one to get started."),
			"",
			WelcomeHelpStyle.Render("[n] New session    [d] Daily session    [q] Quit"),
		)
		return CenterIn(m.width, m.height, content)
	}

	var lines []string
	lines = append(lines, LogoStyle.Render(Logo))
	lines = append(lines, WelcomePromptStyle.Render("Select a session to continue."))
	count := "session"
	if len(m.sessions) != 1 {
		count += "s"
	}
	lines = append(lines, SessionCountStyle.Render(fmt.Sprintf("%d %s", len(m.sessions), count)))
	lines = append(lines, m.sessionsView())
	wf := bubbleshelp.New()
	wf.SetWidth(m.width)
	lines = append(lines, lipgloss.NewStyle().Padding(0, 1).Render(wf.View(welcomeFooter{})))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if m.err != nil {
		content = ErrorBannerStyle.Render("Error: "+m.err.Error()) + "\n" + content
	}
	return CenterIn(m.width, m.height, content)
}

// sessionsView renders the session list without the stock list.Model chrome
// (no title bar, no status bar). Navigation still happens through
// m.sessionsList; this is purely presentational.
func (m Model) sessionsView() string {
	w := m.sessionsList.Width()
	h := m.sessionsList.Height()
	if h < 1 {
		h = 8
	}
	idx := m.sessionsList.Index()
	start := 0
	if idx >= h {
		start = idx + 1 - h
	}
	end := start + h
	if end > len(m.sessions) {
		end = len(m.sessions)
	}

	var rows []string
	for i := start; i < end; i++ {
		name := m.sessions[i].Name
		if lipgloss.Width(name) > w-2 {
			name = ansi.Truncate(name, w-2, "…")
		}
		if i == idx {
			rows = append(rows, SelectedTaskStyle.Bold(true).Width(w).Render(name))
		} else {
			rows = append(rows, TaskStyle.Width(w).Render(name))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// welcomeKeys handles KeyPressMsg on the session menu. Navigation keys that
// Listly does not bind itself are routed to the sessions list.
func (m *Model) welcomeKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.shutdown = true
		return m, tea.Quit
	case "enter":
		if len(m.sessions) > 0 {
			s := m.sessions[m.sessionsList.Index()]
			m.sessionID = s.ID
			m.sessionName = s.Name
			m.setMode(normal)
			m.loadSessionTasks(s.ID)
		}
		return m, nil
	case "?":
		m.returnMode = welcome
		m.setMode(help)
		return m, nil
	case "n", "d", "r", "x":
	default:
		var cmd tea.Cmd
		m.sessionsList, cmd = m.sessionsList.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "n":
		m.returnMode = welcome
		m.setMode(saving)
		m.sessionForm = NewSessionNameForm("Create New Session",
			func(name string) tea.Cmd {
				id, err := m.store.CreateSession(name)
				if err != nil {
					m.err = err
					return nil
				}
				m.err = nil
				m.sessionID = id
				m.sessionName = name
				m.setMode(normal)
				return nil
			},
			func() { m.setMode(m.returnMode) },
		)
		m.sessionForm.width = m.width
		m.sessionForm.height = m.height
		return m, nil
	case "d":
			s, err := m.store.GetOrCreateDailySession()
		if err != nil {
			m.err = err
			return m, nil
		}
		m.err = nil
		m.sessionID = s.ID
		m.sessionName = s.Name
		m.setMode(normal)
		m.loadSessionTasks(s.ID)
	case "r":
		if len(m.sessions) > 0 {
			s := m.sessions[m.sessionsList.Index()]
			m.returnMode = welcome
			m.setMode(saving)
			m.sessionForm = NewSessionNameForm("Rename Session",
				func(name string) tea.Cmd {
					if err := m.store.UpdateSessionName(s.ID, name); err != nil {
						m.err = err
						return nil
					}
					m.err = nil
					if m.sessionID == s.ID {
						m.sessionName = name
					}
					m.setMode(welcome)
					m.toast = &Toast{text: "Session renamed."}
					return tea.Batch(m.loadSessions, toastTick())
				},
				func() { m.setMode(m.returnMode) },
			)
			m.sessionForm.width = m.width
			m.sessionForm.height = m.height
			return m, nil
		}
	case "x":
		if len(m.sessions) > 0 {
			s := m.sessions[m.sessionsList.Index()]
			m.confirm = NewConfirmDialog(
				"Delete session '"+s.Name+"'?",
				func() tea.Cmd {
					m.confirm = nil
					if err := m.store.DeleteSession(s.ID); err != nil {
						m.err = err
						return nil
					}
					m.err = nil
					if m.sessionsList.Index() > 0 && m.sessionsList.Index() >= len(m.sessions)-1 {
						m.sessionsList.Select(m.sessionsList.Index() - 1)
					}
					return m.loadSessions
				},
				func() { m.confirm = nil },
			)
			m.confirm.width = m.width
			m.confirm.height = m.height
		}
	}
	return m, nil
}
