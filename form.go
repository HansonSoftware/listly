package main

import (
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
)

type Form struct {
	title       textinput.Model
	description textarea.Model
	focused     status
	width       int
	height      int
	onCancel    func()
	onToast     func(string)
}

func NewForm(focused status) *Form {
	form := &Form{focused: focused}
	form.title = textinput.New()
	form.title.Placeholder = "Task title"
	form.title.Focus()
	form.title.CharLimit = 80
	form.title.Prompt = ""
	tStyles := textinput.DefaultStyles(true)
	tStyles.Focused.Text = lipgloss.NewStyle().Foreground(ColorFg)
	tStyles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ColorMuted)
	form.title.SetStyles(tStyles)

	form.description = textarea.New()
	form.description.Placeholder = "Description (optional)"
	form.description.CharLimit = 500
	form.description.ShowLineNumbers = false
	dStyles := textarea.DefaultStyles(true)
	dStyles.Focused.Base = FocusedTextAreaStyle
	dStyles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ColorMuted)
	dStyles.Blurred.Base = TextAreaStyle
	dStyles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(ColorMuted)
	form.description.SetStyles(dStyles)

	form.title.SetWidth(54)
	form.description.SetWidth(54)
	form.description.SetHeight(5)

	return form
}

func (m Form) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Form) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.onCancel != nil {
				m.onCancel()
			}
			return nil
		case "enter":
			if strings.TrimSpace(m.title.Value()) == "" {
				if m.onToast != nil {
					m.onToast("Please enter a task name.")
				}
				return toastTick()
			}
			f := *m
			return f.CreateTask
		case "tab", "shift+tab":
			if m.title.Focused() {
				m.title.Blur()
				m.description.Focus()
			} else {
				m.description.Blur()
				m.title.Focus()
			}
			return textarea.Blink
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.title.Focused() {
		m.title, cmd = m.title.Update(msg)
		return cmd
	} else {
		m.description, cmd = m.description.Update(msg)
		return cmd
	}
}

func (m Form) CreateTask() tea.Msg {
	task := NewTask(m.focused, m.title.Value(), m.description.Value())
	return task
}

func (m Form) View() tea.View {
	titleInput := m.title.View()
	descInput := m.description.View()

	help := "Enter: create • Tab: switch • Esc: cancel"

	content := lipgloss.JoinVertical(lipgloss.Left,
		CardTitleStyle.Render("New Task"),
		"",
		"Title:",
		titleInput,
		"",
		"Description:",
		descInput,
		"",
		HelpStyle.Render(help),
	)

	card := CardStyle.Render(content)
	v := tea.NewView(CenterIn(m.width, m.height, card))
	v.AltScreen = true
	v.WindowTitle = "Listly"
	return v
}

// SessionNameForm is a simple form for entering a session name
type SessionNameForm struct {
	title    string
	name     textinput.Model
	onSave   func(string) tea.Cmd
	onCancel func()
	width    int
	height   int
}

func NewSessionNameForm(title string, onSave func(string) tea.Cmd, onCancel func()) *SessionNameForm {
	f := &SessionNameForm{title: title, onSave: onSave, onCancel: onCancel}
	f.name = textinput.New()
	f.name.Placeholder = "Session name"
	f.name.Focus()
	f.name.CharLimit = 50
	f.name.Prompt = ""
	nStyles := textinput.DefaultStyles(true)
	nStyles.Focused.Text = lipgloss.NewStyle().Foreground(ColorFg)
	nStyles.Focused.Placeholder = lipgloss.NewStyle().Foreground(ColorMuted)
	f.name.SetStyles(nStyles)
	f.name.SetWidth(56)
	return f
}

func (m SessionNameForm) Init() tea.Cmd {
	return nil
}

func (m *SessionNameForm) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.onCancel != nil {
				m.onCancel()
			}
			return nil
		case "enter":
			name := m.name.Value()
			if name != "" {
				if m.onSave != nil {
					return m.onSave(name)
				}
			} else if m.onCancel != nil {
				m.onCancel()
			}
			return nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	m.name, cmd = m.name.Update(msg)
	return cmd
}

func (m SessionNameForm) View() tea.View {
	input := m.name.View()

	content := lipgloss.JoinVertical(lipgloss.Left,
		CardTitleStyle.Render(m.title),
		"",
		"Name:",
		input,
		"",
		HelpStyle.Render("Enter: save  •  Esc: cancel"),
	)

	card := CardStyle.Render(content)
	v := tea.NewView(CenterIn(m.width, m.height, card))
	v.AltScreen = true
	v.WindowTitle = "Listly"
	return v
}
