package main

import (
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
)

type status int

const (
	todo       status = 0
	completing        = 1
	done              = 2
)

type mode int

const (
	welcome   mode = 0
	normal         = 1
	creation       = 2
	filtering      = 3
	saving         = 4
	help           = 5
)

type Model struct {
	lists        []list.Model
	histories    histories
	focused      status
	loaded       bool
	shutdown     bool
	mode         mode
	err          error
	sessionID    int64
	sessionName  string
	toast        *Toast
	sessions     []Session
	sessionsList list.Model
	confirm      *ConfirmDialog
	width        int
	height       int
	store        Store
	form         *Form
	sessionForm  *SessionNameForm
	returnMode   mode
}

func New(store Store) *Model {
	// Lists start with safe defaults; recreateLists replaces them with
	// real items/dimensions on the first WindowSizeMsg.
	lists := make([]list.Model, 3)
	for i := range lists {
		lists[i] = list.New(nil, TaskDelegate{}, 40, 20)
	}

	sl := list.New(nil, SessionDelegate{}, lipgloss.Width(Logo), 8)
	// Ditch the charm list's left/right pagination keys; up/down/j/k and
	// g/G are the only navigation we want.
	sl.KeyMap.NextPage.Unbind()
	sl.KeyMap.PrevPage.Unbind()
	sl.Title = "Welcome!"
	sl.SetStatusBarItemName("session", "sessions")
	sl.SetShowHelp(false)
	// Left-align the title bar (title + item count) with the logo/footer.
	sl.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 1).Margin(0)
	return &Model{
		mode:         welcome,
		lists:        lists,
		sessionsList: sl,
		histories:    histories{},
		store:        store,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadSessions
}

func (m *Model) loadSessions() tea.Msg {
	sessions, err := m.store.ListSessions()
	if err != nil {
		return err
	}
	return sessionsLoadedMsg{sessions}
}

type sessionsLoadedMsg struct {
	sessions []Session
}

// setMode is the single path for switching screens/modes.
func (m *Model) setMode(md mode) {
	m.mode = md
}

// history returns the undo/redo history for the active session.
func (m *Model) history() *History {
	return m.histories.get(m.sessionID)
}

func (m *Model) pushUndo(undoFn, redoFn func()) {
	h := m.history()
	h.undo.Push(undoFn, redoFn)
	// A fresh action invalidates any pending redo.
	h.redo.Clear()
}

func (m *Model) undo() {
	h := m.history()
	if o, ok := h.undo.Pop(); ok {
		if o.undo != nil {
			o.undo()
		}
		h.redo.Push(o.undo, o.redo)
		m.autoSave()
	}
}

func (m *Model) redo() {
	h := m.history()
	if o, ok := h.redo.Pop(); ok {
		if o.redo != nil {
			o.redo()
		}
		h.undo.Push(o.undo, o.redo)
		m.autoSave()
	}
}

func (m *Model) autoSave() {
	if m.sessionID == 0 {
		return
	}
	tasks := m.getAllTasks()
	m.err = m.store.SaveSession(m.sessionID, tasks)
}

func (m *Model) getAllTasks() []Task {
	var tasks []Task
	for _, l := range m.lists {
		for _, item := range l.Items() {
			tasks = append(tasks, item.(Task))
		}
	}
	return tasks
}

func (m Model) View() tea.View {
	v := m.view()
	if m.toast != nil {
		x := m.width - lipgloss.Width(m.toast.Render()) - 1
		if x < 0 {
			x = 0
		}
		v.Content = overlay(v.Content, m.toast.Render(), x, 1)
	}
	return v
}

func (m Model) view() tea.View {
	switch m.mode {
	case welcome:
		if m.confirm != nil {
			return m.confirm.View()
		}
		v := tea.NewView(m.welcomeView())
		v.AltScreen = true
		v.WindowTitle = "Listly"
		return v
	case saving:
		if m.sessionForm != nil {
			v := m.sessionForm.View()
			v.AltScreen = true
			v.WindowTitle = "Listly"
			return v
		}
		return tea.NewView("")
	case creation:
		if m.form != nil {
			v := m.form.View()
			v.AltScreen = true
			v.WindowTitle = "Listly"
			return v
		}
		return tea.NewView("")
	case help:
		v := tea.NewView(m.helpView())
		v.AltScreen = true
		v.WindowTitle = "Listly"
		return v
	case normal, filtering:
		if m.confirm != nil {
			return m.confirm.View()
		}
		v := tea.NewView(m.mainView())
		v.AltScreen = true
		v.WindowTitle = "Listly"
		return v
	default:
		return tea.NewView("")
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recreateLists()
		m.loaded = true
		w := lipgloss.Width(Logo)
		if w > m.width-2 {
			w = m.width - 2
		}
		h := m.height - lipgloss.Height(Logo) - 6
		if h < 3 {
			h = 3
		}
		m.sessionsList.SetSize(w, h)
		if m.form != nil && m.mode == creation {
			m.form.width = msg.Width
			m.form.height = msg.Height
		}
		if m.sessionForm != nil && m.mode == saving {
			m.sessionForm.width = msg.Width
			m.sessionForm.height = msg.Height
		}
		if m.confirm != nil {
			m.confirm.width = msg.Width
			m.confirm.height = msg.Height
		}

	case sessionsLoadedMsg:
		m.err = nil // successful reload clears any earlier persistence error
		m.sessions = msg.sessions
		items := make([]list.Item, 0, len(msg.sessions))
		for _, s := range msg.sessions {
			items = append(items, s)
		}
		m.sessionsList.SetItems(items)
		m.sessionsList.Select(0)

	case toastMsg:
		m.toast = nil

	case error:
		m.err = msg // loadSessions surfaces store errors as msgs

	case tea.KeyPressMsg:
		if m.confirm != nil {
			return m, m.confirm.Update(msg)
		}
		switch m.mode {
		case creation:
			if m.form != nil {
				return m, m.form.Update(msg)
			}
		case saving:
			if m.sessionForm != nil {
				return m, m.sessionForm.Update(msg)
			}
		case welcome:
			return m.welcomeKeys(msg)
		case normal:
			if _, cmd, handled := m.normalKeys(msg); handled {
				return m, cmd
			}
		case filtering:
			return m.filteringKeys(msg)
		case help:
			return m.helpKeys(msg)
		}

	case Task:
		return m.handleNewTask(msg)
	}

	// Update focused list, but only when a board is visible — forwarding
	// keys to a list while in welcome/help/saving/creation modes would
	// silently mutate the board, and updating before recreateLists ran
	// panicked on the zero-value list in v2.
	if m.loaded && (m.mode == normal || m.mode == filtering) {
		var cmd tea.Cmd
		m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
