package main

import (
	"github.com/charmbracelet/bubbles/list"
	css "github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	lists    []list.Model
	focused  status
	loaded   bool
	shutdown bool
	err      error
}

func New() *Model {
	return &Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
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

func (m Model) View() string {
	// NOTE: Eliminates final std out print, leaving terminal clean.
	if m.shutdown {
		return ""
	}

	if m.loaded {
		todoView := m.lists[todo].View()
		completingView := m.lists[completing].View()
		doneView := m.lists[done].View()

		switch m.focused {

		case completing:
			return css.JoinHorizontal(
				css.Top,
				columnStyle.Render(todoView),
				focusedStyle.Render(completingView),
				columnStyle.Render(doneView),
			)

		case done:
			return css.JoinHorizontal(
				css.Top,
				columnStyle.Render(todoView),
				columnStyle.Render(completingView),
				focusedStyle.Render(doneView),
			)

		default:
			return css.JoinHorizontal(
				css.Top,
				focusedStyle.Render(todoView),
				columnStyle.Render(completingView),
				columnStyle.Render(doneView),
			)
		}

	} else {
		// TODO: Loading spinner
		return "Loading ..."
	}
}

func (m *Model) CreateLists(width int, height int) {
	d := list.New([]list.Item{}, list.NewDefaultDelegate(), width/3, height-4)
	d.SetShowHelp(false)

	m.lists = []list.Model{d, d, d}

	// Todo list
	m.lists[todo].Title = "Today's Agenda"
	m.lists[todo].SetItems([]list.Item{
		Task{status: todo, title: "complete c# training", description: "update cs-neetcode repo"},
		Task{status: todo, title: "lunch", description: "sushi @ 12:00pm"},
	})

	// Completing list
	m.lists[completing].Title = "Working On"
	m.lists[completing].SetItems([]list.Item{
		Task{status: completing, title: "implement client feedback", description: "substation modeling"},
	})

	// Done list
	m.lists[done].Title = "Done"
	m.lists[done].SetItems([]list.Item{
		Task{status: done, title: "meeting @ 9:00am", description: "engineering team stand-up"},
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if !m.loaded {
			columnStyle.Width(msg.Width / 3)
			columnStyle.Height(msg.Height - 4)
			focusedStyle.Width(msg.Width / 3)
			focusedStyle.Height(msg.Height - 4)
			m.CreateLists(msg.Width, msg.Height)
			m.loaded = true
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.shutdown = true
			return m, tea.Quit
		case "left", "h", "shift+tab":
			m.Prev()
		case "right", "l", "tab":
			m.Next()
		}
	}

	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)

	return m, cmd
}
