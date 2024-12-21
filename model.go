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

var models []tea.Model

const (
	model status = iota
	form
)

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

func (m *Model) DeleteTask() tea.Msg {
	if m.lists[m.focused].SelectedItem() != nil {
		selectedItem := m.lists[m.focused].SelectedItem()
		selectedTask := selectedItem.(Task)
		m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
		return nil
	}
	return nil
}

func (m *Model) MoveToNext() tea.Msg {
	if m.lists[m.focused].SelectedItem() != nil {
		selectedItem := m.lists[m.focused].SelectedItem()
		selectedTask := selectedItem.(Task)
		m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
		selectedTask.Next()
		m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items())-1, list.Item(selectedTask))
		return nil
	}
	return nil
}

func (m Model) View() string {
	var cols []string
	if m.shutdown {
		return ""
	}

	if m.loaded {
		todoView := m.lists[todo].View()
		completingView := m.lists[completing].View()
		doneView := m.lists[done].View()

		switch m.focused {

		case completing:
			cols = []string{
				columnStyle.Render(todoView),
				focusedStyle.Render(completingView),
				columnStyle.Render(doneView),
			}

		case done:
			cols = []string{
				columnStyle.Render(todoView),
				columnStyle.Render(completingView),
				focusedStyle.Render(doneView),
			}

		default:
			cols = []string{
				focusedStyle.Render(todoView),
				columnStyle.Render(completingView),
				columnStyle.Render(doneView),
			}
		}

		return css.JoinHorizontal(css.Left, cols...)
	} else {
		return "Loading..."
	}
}

func (m *Model) CreateLists(width int, height int) {
	d := list.New([]list.Item{}, list.NewDefaultDelegate(), width/4, height-4*2)
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
			columnStyle.Width(msg.Width)
			focusedStyle.Width(msg.Width)
			m.CreateLists(msg.Width, msg.Height)
			m.loaded = true
		}

	case tea.KeyMsg:
		// TODO: Have cases for different modes
		// i.e. when you are filling out a task, q should not quit
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
		case "n":
			models[model] = m
			models[form] = NewForm(m.focused)
			return models[form].Update(nil)
		}
	case Task:
		task := msg
		return m, m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
	}

	var cmd tea.Cmd
	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)

	return m, cmd
}
