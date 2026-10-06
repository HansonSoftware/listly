package main

import (
	bubbleshelp "charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
)

func boardShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("tab", "h", "l"), key.WithHelp("←/→/tab", "columns")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "move")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
		key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "delete")),
		key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "undo")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "sessions")),
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more")),
		key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

// boardFooter implements help.KeyMap so the shared help component can
// render the board footer.
type boardFooter struct{}

func (boardFooter) ShortHelp() []key.Binding  { return boardShortHelp() }
func (boardFooter) FullHelp() [][]key.Binding { return [][]key.Binding{boardShortHelp()} }

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

func (m *Model) DeleteTask() tea.Msg {
	if m.lists[m.focused].SelectedItem() != nil {
		l := &m.lists[m.focused]
		item := l.SelectedItem()
		task := item.(Task)
		idx := l.Index()
		l.RemoveItem(idx)
		m.pushUndo(
			func() {
				m.lists[task.status].InsertItem(idx, item)
				m.lists[task.status].Select(idx)
			},
			func() {
				m.lists[task.status].RemoveItem(idx)
			},
		)
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
		m.pushUndo(
			func() {
				m.lists[after.status].RemoveItem(newIdx)
				m.lists[before.status].InsertItem(idx, list.Item(before))
				m.lists[before.status].Select(idx)
			},
			func() {
				m.lists[before.status].RemoveItem(idx)
				m.lists[after.status].InsertItem(newIdx, list.Item(after))
				m.lists[after.status].Select(newIdx)
			},
		)
		m.autoSave()
		return nil
	}
	return nil
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
		m.lists[i].SetShowTitle(false) // no title bar; our column titles sit outside
		m.lists[i].SetFilteringEnabled(true)
		m.lists[i].Title = ""
		// Ditch the charm list's left/right pagination keys; up/down/j/k and
		// g/G are the only navigation we want.
		m.lists[i].KeyMap.NextPage.Unbind()
		m.lists[i].KeyMap.PrevPage.Unbind()
	}
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

	title := BoardTitleStyle.Render(m.sessionName)

	hf := bubbleshelp.New()
	hf.SetWidth(m.width)
	footer := lipgloss.NewStyle().Padding(0, 1).Render(hf.View(boardFooter{}))

	parts := []string{title, board, footer}
	if m.err != nil {
		parts = append([]string{ErrorBannerStyle.Render("Error: " + m.err.Error())}, parts...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m *Model) loadSessionTasks(sessionID int64) {
	tasks, err := m.store.LoadSession(sessionID)
	if err != nil {
		m.err = err
		return
	}
	for i := range m.lists {
		m.lists[i].SetItems(nil)
	}
	for _, task := range tasks {
		m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
	}
}

// normalKeys handles KeyPressMsg while a session board is visible.
// The third return value reports whether the key was consumed; when it is
// false the shared tail in Model.Update forwards the key to the focused
// list (e.g. "j" scrolls the column).
func (m *Model) normalKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.shutdown = true
		return m, tea.Quit, true
	case "left", "h", "shift+tab":
		m.Prev()
	case "esc":
		m.autoSave()
		m.mode = welcome
		return m, m.loadSessions, true
	case "right", "l", "tab":
		m.Next()
	case "enter":
		m.MoveToNext()
	case "x":
		if m.lists[m.focused].SelectedItem() != nil {
			task := m.lists[m.focused].SelectedItem().(Task)
			m.confirm = NewConfirmDialog(
				"Delete task '"+task.Title()+"'?",
				func() tea.Cmd {
					m.confirm = nil
					m.DeleteTask()
					return nil
				},
				func() { m.confirm = nil },
			)
			m.confirm.width = m.width
			m.confirm.height = m.height
		}
	case "?":
		m.returnMode = normal
		m.mode = help
	case "/":
		m.mode = filtering
		m.lists[m.focused].SetFilterState(list.Filtering)
	case "n":
		f := NewForm(m.focused)
		f.width = m.width
		f.height = m.height
		f.onCancel = func() { m.mode = normal }
		f.onToast = func(text string) { m.toast = &Toast{text: text} }
		m.form = f
		m.mode = creation
		return m, m.form.Init(), true
	case "u":
		m.undo()
	case "r":
		m.redo()
	case "ctrl+s":
		m.autoSave()
		m.toast = &Toast{text: "Session saved."}
		return m, toastTick(), true
	default:
		return m, nil, false
	}
	return m, nil, false
}

func (m *Model) filteringKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" || msg.String() == "enter" {
		// Let the list see the key so it can apply or cancel the
		// filter, then return to normal mode.
		m.lists[m.focused], _ = m.lists[m.focused].Update(msg)
		m.mode = normal
		return m, nil
	}
	m.lists[m.focused], _ = m.lists[m.focused].Update(msg)
	if m.lists[m.focused].FilterState() != list.Filtering {
		m.mode = normal
	}
	return m, nil
}

// handleNewTask inserts a task emitted by the creation form and registers
// the matching undo/redo pair.
func (m *Model) handleNewTask(msg Task) (tea.Model, tea.Cmd) {
	task := msg
	idx := len(m.lists[task.status].Items())
	removeAt := func() {
		items := m.lists[task.status].Items()
		if idx < len(items) {
			if t, ok := items[idx].(Task); ok && t.Title() == task.Title() {
				m.lists[task.status].RemoveItem(idx)
			}
		}
	}
	m.pushUndo(
		removeAt,
		func() {
			items := m.lists[task.status].Items()
			if idx <= len(items) {
				m.lists[task.status].InsertItem(idx, task)
				m.lists[task.status].Select(idx)
			}
		},
	)
	if m.mode == creation {
		m.mode = normal
	}
	return m, m.lists[task.status].InsertItem(idx, task)
}
