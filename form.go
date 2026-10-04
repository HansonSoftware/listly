package main

import (
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

	form.title.SetWidth(56)
	form.description.SetWidth(56)
	form.description.SetHeight(5)

	return form
}

func (m Form) Init() tea.Cmd {
	return textarea.Blink
}

func (m Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return models[model], nil
		case "enter":
			if m.title.Focused() {
				m.title.Blur()
				m.description.Focus()
				return m, textarea.Blink
			} else {
				models[form] = m
				return models[model], m.CreateTask
			}
		case "tab", "shift+tab":
			// Prevent tab from switching columns while in form
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.title.Focused() {
		m.title, cmd = m.title.Update(msg)
		return m, cmd
	} else {
		m.description, cmd = m.description.Update(msg)
		return m, cmd
	}
}

func (m Form) CreateTask() tea.Msg {
	task := NewTask(m.focused, m.title.Value(), m.description.Value())
	return task
}

func (m Form) View() tea.View {
	titleInput := m.title.View()
	descInput := m.description.View()

	focusedField := "title"
	if m.description.Focused() {
		focusedField = "description"
	}

	var help string
	if focusedField == "title" {
		help = "Enter: next field  •  Esc: cancel"
	} else {
		help = "Enter: create task  •  Esc: cancel"
	}

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
	name   textinput.Model
	onSave func(string) (tea.Model, tea.Cmd)
	width  int
	height int
}

func NewSessionNameForm(onSave func(string) (tea.Model, tea.Cmd)) *SessionNameForm {
	f := &SessionNameForm{onSave: onSave}
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

func (m SessionNameForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return models[model], nil
		case "enter":
			name := m.name.Value()
			if name != "" {
				return m.onSave(name)
			}
			return models[model], nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	m.name, cmd = m.name.Update(msg)
	return m, cmd
}

func (m SessionNameForm) View() tea.View {
	input := m.name.View()

	content := lipgloss.JoinVertical(lipgloss.Left,
		CardTitleStyle.Render("Save Session"),
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
