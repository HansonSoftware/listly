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
	lists      []list.Model
	undoStack *UndoStack
	focused    status
	loaded     bool
	shutdown   bool
	mode       mode
	err        error
	sessionID  int64
	isDaily    bool
	welcomeIdx int
	sessions   []Session
	width      int
	height     int
	store      Store
	form       *Form
	sessionForm *SessionNameForm
	returnMode mode
}

func New(store Store) *Model {
	return &Model{
		mode:  welcome,
		lists: make([]list.Model, 3),
		undoStack: NewUndoStack(50),
		store: store,
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

func (m *Model) Next() {
	if m.focused == done {
		m.focused = todo
	} else {
		m.focused++
	}
}

func (m *Model) Prev() {
	if m.focused == todo {
		m.focused = done
	} else {
		m.focused--
	}
}

func (m *Model) pushUndo(fn func()) {
	m.undoStack.Push(fn)
}

func (m *Model) undo() {
	if fn := m.undoStack.Pop(); fn != nil {
		fn()
	}
}

func (m *Model) DeleteTask() tea.Msg {
	if m.lists[m.focused].SelectedItem() != nil {
		l := &m.lists[m.focused]
		item := l.SelectedItem()
		task := item.(Task)
		idx := l.Index()
		l.RemoveItem(idx)
		m.pushUndo(func() {
			m.lists[task.status].InsertItem(idx, item)
			m.lists[task.status].Select(idx)
				m.autoSave()
		})
			m.autoSave()
		return nil
	}
	return nil
}

func (m *Model) MoveToNext() tea.Msg {
	if m.lists[m.focused].SelectedItem() != nil {
		selectedItem := m.lists[m.focused].SelectedItem()
		before := selectedItem.(Task)
		idx := m.lists[m.focused].Index()
		after := before
		after.Next()
		m.lists[before.status].RemoveItem(idx)
		newIdx := len(m.lists[after.status].Items())
		m.lists[after.status].InsertItem(newIdx, list.Item(after))
		m.pushUndo(func() {
			m.lists[after.status].RemoveItem(newIdx)
			m.lists[before.status].InsertItem(idx, list.Item(before))
			m.lists[before.status].Select(idx)
				m.autoSave()
		})
			m.autoSave()
		return nil
	}
	return nil
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

func (m *Model) recreateLists() {
	if m.width == 0 || m.height == 0 {
		return
	}
	if len(m.lists) != 3 {
		m.lists = make([]list.Model, 3)
	}
	layout := ColumnLayout(m.width, m.height)

	delegate := TaskDelegate{}

	for i := range m.lists {
		items := m.lists[i].Items()
		m.lists[i] = list.New(items, delegate, layout.ColInternalWidth, layout.ColHeight)
		m.lists[i].SetShowHelp(false)
		m.lists[i].SetShowStatusBar(false)
		m.lists[i].SetFilteringEnabled(false)
		m.lists[i].Title = ""
	}
}

func (m Model) View() tea.View {
	switch m.mode {
	case welcome:
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
		v := tea.NewView(m.mainView())
		v.AltScreen = true
		v.WindowTitle = "Listly"
		return v
	default:
		return tea.NewView("")
	}
}

func (m Model) welcomeView() string {
	if len(m.sessions) == 0 {
		content := lipgloss.JoinVertical(lipgloss.Center,
			WelcomeTitleStyle.Render("Welcome to Listly"),
			WelcomeSubtitleStyle.Render("No sessions yet. Create one to get started."),
			"",
			WelcomeHelpStyle.Render("[n] New session    [d] Daily session    [q] Quit"),
		)
		return CenterIn(m.width, m.height, content)
	}

	var lines []string
	lines = append(lines, WelcomeTitleStyle.Render("Welcome to Listly"))
	lines = append(lines, WelcomeSubtitleStyle.Render("Select a session to continue"))
	lines = append(lines, "")

	for i, s := range m.sessions {
		var line string
		dailyMark := ""
		if s.IsDaily {
			dailyMark = DailySessionBadge.Render("(daily)")
		}
		name := s.Name + dailyMark

		if i == m.welcomeIdx {
			line = SelectedSessionStyle.Render("▸ " + name)
		} else {
			line = SessionItemStyle.Render("  " + name)
		}
		lines = append(lines, line)
	}

	lines = append(lines, "")
	lines = append(lines, WelcomeHelpStyle.Render("[↑/↓] Navigate  [enter] Open  [n] New  [d] Daily  [q] Quit"))

	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if m.err != nil {
		content = ErrorBannerStyle.Render("Error: "+m.err.Error()) + "\n" + content
	}
	return CenterIn(m.width, m.height, content)
}

func (m Model) mainView() string {
	if m.shutdown {
		return ""
	}

	if !m.loaded {
		return CenterIn(m.width, m.height, "Loading...")
	}

	titles := []string{
		"Today's Agenda",
		"Working On",
		"Done",
	}

	layout := ColumnLayout(m.width, m.height)

	var cols []string
	for i := 0; i < 3; i++ {
		view := m.lists[i].View()

		title := titles[i]
		if i == int(m.focused) {
			title = FocusedColumnTitleStyle.Render("▸ " + title)
		} else {
			title = ColumnTitleStyle.Render("  " + title)
		}

		content := lipgloss.JoinVertical(lipgloss.Left, title, view)
		content = lipgloss.NewStyle().MaxWidth(layout.ColContentWidth).Render(content)

		if i == int(m.focused) {
			cols = append(cols, FocusedColumnStyle.Width(layout.ColTotalWidth).Render(content))
		} else {
			cols = append(cols, ColumnStyle.Width(layout.ColTotalWidth).Render(content))
		}
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	board = lipgloss.NewStyle().MarginTop(1).Render(board)

	help := HelpStyle.Width(m.width).Render("←/→/Tab: switch columns  •  Enter: move task  •  n: new  •  d: delete  •  u: undo  •  Ctrl+s: save  •  /: filter  •  ?: help  •  q: quit")

	parts := []string{board, help}
	if m.err != nil {
		parts = append([]string{ErrorBannerStyle.Render("Error: " + m.err.Error())}, parts...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) helpView() string {
	keybinds := []struct {
		key  string
		desc string
	}{
		{"h / ← / Shift+Tab", "Previous column"},
		{"l / → / Tab", "Next column"},
		{"Enter", "Move task to next column"},
		{"d", "Delete selected task"},
		{"n", "New task"},
		{"u", "Undo last action"},
		{"Ctrl+s", "Save session"},
		{"/", "Filter mode"},
		{"?", "Show this help"},
		{"q / Ctrl+c", "Quit"},
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

func (m *Model) loadSessionTasks(sessionID int64) {
	tasks, err := m.store.LoadSession(sessionID)
	if err != nil {
		m.err = err
		return
	}
	for _, task := range tasks {
		m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.loaded {
			m.recreateLists()
			m.loaded = true
		} else {
			m.recreateLists()
		}
		if m.form != nil && m.mode == creation {
			m.form.width = msg.Width
			m.form.height = msg.Height
		}
		if m.sessionForm != nil && m.mode == saving {
			m.sessionForm.width = msg.Width
			m.sessionForm.height = msg.Height
		}

	case sessionsLoadedMsg:
		m.sessions = msg.sessions
		m.welcomeIdx = 0

	case tea.KeyPressMsg:
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
			switch msg.String() {
			case "ctrl+c", "q":
				m.shutdown = true
				return m, tea.Quit
			case "up", "k":
				if m.welcomeIdx > 0 {
					m.welcomeIdx--
				}
			case "down", "j":
				if m.welcomeIdx < len(m.sessions)-1 {
					m.welcomeIdx++
				}
			case "enter":
				if len(m.sessions) > 0 {
					s := m.sessions[m.welcomeIdx]
					m.sessionID = s.ID
					m.isDaily = s.IsDaily
					m.mode = normal
					m.loadSessionTasks(s.ID)
				}
			case "n":
				m.returnMode = welcome
				m.mode = saving
				m.sessionForm = NewSessionNameForm(
					func(name string) {
						id, err := m.store.CreateSession(name)
						if err != nil {
							m.err = err
							return
						}
						m.sessionID = id
						m.isDaily = false
						m.mode = normal
					},
					func() { m.mode = m.returnMode },
				)
				return m, nil
			case "d":
				id, err := m.store.GetDailySession()
				if err != nil {
					m.err = err
					return m, nil
				}
				m.sessionID = id
				m.isDaily = true
				m.mode = normal
				m.loadSessionTasks(id)
			}

		case normal:
			switch msg.String() {
			case "ctrl+c", "q":
				m.shutdown = true
				return m, tea.Quit
			case "left", "h", "shift+tab":
				m.Prev()
			case "right", "l", "tab":
				m.Next()
			case "enter":
				m.MoveToNext()
			case "d":
				m.DeleteTask()
			case "?":
				m.mode = help
			case "/":
				m.mode = filtering
			case "n":
				f := NewForm(m.focused)
				f.width = m.width
				f.height = m.height
				f.onCancel = func() { m.mode = normal }
				m.form = f
				m.mode = creation
				return m, m.form.Init()
			case "u":
				m.undo()
					m.autoSave()
			case "ctrl+s":
				m.returnMode = normal
				m.mode = saving
				m.sessionForm = NewSessionNameForm(
					func(name string) {
						if m.sessionID == 0 {
							id, err := m.store.CreateSession(name)
							if err != nil {
								m.err = err
								return
							}
							m.sessionID = id
						} else {
							// Update existing session name
							err := m.store.UpdateSessionName(m.sessionID, name)
							if err != nil {
								m.err = err
								return
							}
						}
						m.mode = normal
						m.autoSave()
					},
					func() { m.mode = m.returnMode },
				)
				return m, nil
			}

		case filtering:
			switch msg.String() {
			case "esc", "enter":
				m.mode = normal
			}

		case help:
			switch msg.String() {
			case "?", "esc", "enter":
				m.mode = normal
				return m, nil
			case "ctrl+c", "q":
				m.shutdown = true
				return m, tea.Quit
			}
		}

	case Task:
		task := msg
		idx := len(m.lists[task.status].Items())
		m.pushUndo(func() {
			items := m.lists[task.status].Items()
			if idx < len(items) {
				if t, ok := items[idx].(Task); ok && t.Title() == task.Title() {
					m.lists[task.status].RemoveItem(idx)
				}
			}
				m.autoSave()
		})
		if m.mode == creation {
			m.mode = normal
		}
		return m, m.lists[task.status].InsertItem(idx, task)
	}

	// Update focused list
	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
